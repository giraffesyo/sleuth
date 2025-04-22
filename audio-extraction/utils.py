import time


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
            start = convert_timestamp_to_hhmmss(segment["start"])
            end = convert_timestamp_to_hhmmss(segment["end"])
            f.write(f"{start} - {end}: {segment['text']}\n")
    print(f"Relevant transcription saved to {transcript_path}")
