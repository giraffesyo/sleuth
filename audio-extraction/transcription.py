import whisper
import ffmpeg
import os
import time
import requests
import json

from dotenv import load_dotenv

load_dotenv()


def extract_audio(video_path, audio_path):
    """Extract audio from video file

    Args:
        video_path (file): Path to video file
        audio_path (file): Path to save audio file
    """
    try:
        input_video = ffmpeg.input(video_path)
        output_audio = ffmpeg.output(input_video, audio_path)
        ffmpeg.run(output_audio)
        return output_audio_path
    except Exception as e:
        print("Error extracting audio: ", e)
        return None


def transcribe_audio(audio_path):
    """Transcribe audio file to text

    Args:
        audio_path (file): Path to audio file
    """
    try:
        model = whisper.load_model("base")
        audio = model.transcribe(audio_path)

        return audio["text"], audio["segments"]
    except Exception as e:
        print("Error transcribing audio: ", e)
        return None


def find_relevant_timestamps(segments, keywords):
    """Find relevant timestamps for keywords

    Args:
        segments (list): List of audio segments
        keywords (list): List of keywords
    """
    relevant_timestamps = []
    for segment in segments:
        text = segment["text"].lower()
        if any(keyword in text for keyword in keywords):
            # save times as 2 decimal places
            relevant_timestamps.append(
                {
                    "start": round(segment["start"], 2),
                    "end": round(segment["end"], 2),
                    "text": segment["text"],
                }
            )
    return relevant_timestamps


def llama3(prompt):
    """Send prompt to LLaMA-3.1 API and return response
    Currently using Llama3.1 - 8B model

    Args:
        prompt (str): Prompt to send to LLaMA-3.1

    Returns:
        str: Response from LLaMA-3.1
    """
    # url = "http://localhost:11434/api/chat"
    url = os.getenv("LLAMA3_URL")
    if not url:
        print("LLAMA3_URL environment variable not set")
        return None

    data = {
        "model": "llama3.1",
        "messages": [{"role": "user", "content": prompt}],
        "stream": False,
    }

    headers = {"Content-Type": "application/json"}

    try:
        response = requests.post(url, headers=headers, json=data)
        return response.json()["message"]["content"]
    except Exception as e:
        print(f"Error with LLaMA-3: {e}")
        return None


def detect_relevant_timestamps_with_llama(segments):
    """Use LLaMA-3 to find timestamps where body discovery is discussed.

    Args:
        transcript (str): Full transcribed text - not used for now.
        segments (list): List of transcribed segments with timestamps.

    Returns:
        list: Relevant timestamps with extracted text.
    """
    relevant_timestamps = []
    for segment in segments:
        text = segment["text"]
        prompt = f"Does this text discuss any crime scene, or body found, or dead people? Reply with 'Yes' or 'No'.\nText: {text}"
        print(f"Prompt: {prompt}")
        try:
            response = llama3(prompt)
            print(f"Response: {response}")
            if "Yes" in response:
                relevant_timestamps.append(
                    {
                        "start": round(segment["start"], 2),
                        "end": round(segment["end"], 2),
                        "text": segment["text"],
                    }
                )
        except Exception as e:
            print(f"Error detecting relevant timestamps: {e}")
            continue
    return relevant_timestamps


def convert_timestamp_to_hhmmss(seconds):
    """Convert seconds to MM:SS format

    Args:
        seconds (int): Seconds to convert
    """
    return time.strftime("%M:%S", time.gmtime(seconds))


def save_keywords_timestamps(relevant_segments, transcript_path):
    """Save relevant timestamps to file

    Args:
        relevant_segments (list): List of relevant timestamps
        transcript_path (file): Path to save transcript file
    """
    with open(transcript_path, "w") as f:
        for segment in relevant_segments:
            # f.write(f"{segment['start']} - {segment['end']}: {segment['text']}\n")
            f.write(
                f"{convert_timestamp_to_hhmmss(segment['start'])} - {convert_timestamp_to_hhmmss(segment['end'])}: {segment['text']}\n"
            )
    print(f"Relevant transcription saved to {transcript_path}")


# Example usage
video_path = "output.mp4"  # Replace with path to video file
transcript_path = "output/transcription_results.txt"
keywords = [
    # "murder",
    # "DNA",
    # "dna",
    "body found",
    "dead",
    "body",
    "murderer",
    "victim",
    "crime scene",
    # "criminal",
    "location of the body",
    "discovery site",
]  # Keywords to search for in transcript

# audio_path = extract_audio(video_path)
# transcript, segments = transcribe_audio(audio_path)

transcript, segments = transcribe_audio(video_path)

# relevant_segments = find_relevant_timestamps(segments, keywords)

# test with LLama
relevant_segments = detect_relevant_timestamps_with_llama(segments)

# Save output
save_keywords_timestamps(relevant_segments, transcript_path)
