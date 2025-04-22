import os
import requests
import json

from utils import convert_timestamp_to_hhmmss

from dotenv import load_dotenv

load_dotenv()


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


def detect_body_discovery_events_full_context(segments):
    """Use LLaMA-3 to find timestamps where body discovery is discussed.

    Args:
        segments (list): List of transcribed segments with timestamps.
    Returns:
        json: List of relevant timestamps with extracted text and other details.
    """
    # Format full text with timestamps
    formatted_text = ""
    for segment in segments:
        start = convert_timestamp_to_hhmmss(segment["start"])
        end = convert_timestamp_to_hhmmss(segment["end"])
        formatted_text += f"{start} - {end}: {segment['text']}\n"

    # Send full text to LLaMA-3
    prompt = f"""
    You are an investigator assistant AI.

    Given this transcript from a video, extract **all mentions** of body discoveries, crime scenes, or similar events.

    For each mention, return:
    - `start`: Start time of the segment (MM:SS)
    - `end`: End time of the segment (MM:SS)
    - `text_snippet`: The snippet where the mention occurs
    - `location`: If mentioned, where the body was found
    - `time_detail`: Any info on *when* it happened (if stated)

    Respond in the following JSON format:
    [
    {{
        "start": "MM:SS",
        "end": "MM:SS",
        "text_snippet": "...",
        "location": "...",
        "time_detail": "..."
    }},
    ...
    ]
    
    Respond with **only** the JSON array, with no additional text, commentary, or explanations.

    Transcript:
    {formatted_text}
    """

    response = llama3(prompt)
    try:
        # print("LLaMA Structured Response:\n", response)
        # Try to parse response safely
        structured_results = json.loads(response)
        return structured_results
    except json.JSONDecodeError:
        # if the model still wraps it, try to strip
        start = response.find("[")
        end = response.rfind("]")
        return json.loads(response[start : end + 1])
    except Exception as e:
        print("Error parsing LLaMA structured output:", e)
        return []
