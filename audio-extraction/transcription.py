import whisper
import ffmpeg
import os


def extract_audio(video_path, audio_path):
    """Extract audio from video file

    Args:
        video_path (file): Path to video file
        audio_path (file): Path to save audio file
    """
    # command = f"ffmpeg -i {video_path} -ar 16000 -ac 1 -c:a pcm_s16le {audio_path}"
    # # -ar 16000: Set audio rate to 16000 Hz
    # # -ac 1: Set audio channels to 1 (mono)

    # os.system(command)
    try:
        input_video = ffmpeg.input(video_path)
        # , c:a='pcm_s16le' , ar=16000, ac=1
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
        segments (list): List
        keywords (list): List of keywords
    """
    relevant_timestamps = []
    for segment in segments:
        text = segment["text"].lower()
        if any(keyword in text for keyword in keywords):
            relevant_timestamps.append(
                {
                    "start": segment["start"],
                    "end": segment["end"],
                    "text": segment["text"],
                }
            )
    return relevant_timestamps


def save_keywords_timestamps(relevant_segments, transcript_path):
    """Save relevant timestamps to file

    Args:
        relevant_segments (list): List of relevant timestamps
        transcript_path (file): Path to save transcript file
    """
    with open(transcript_path, "w") as f:
        for segment in relevant_segments:
            f.write(f"{segment['start']} - {segment['end']}: {segment['text']}\n")
    print(f"Relevant transcription saved to {transcript_path}")


# Example usage
video_path = "output.mp4"  # Replace with path to video file
transcript_path = "output/transcription_results.txt"
keywords = [
    "murder",
    "DNA",
    "dna",
    "body found",
    "location of the body",
    "discovery site",
]  # Keywords to search for in transcript

# audio_path = extract_audio(video_path)
# transcript, segments = transcribe_audio(audio_path)

transcript, segments = transcribe_audio(video_path)

relevant_segments = find_relevant_timestamps(segments, keywords)

# Save output
save_keywords_timestamps(relevant_segments, transcript_path)
