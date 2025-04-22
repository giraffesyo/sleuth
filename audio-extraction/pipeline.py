import json
import os

from transcription import transcribe_audio, extract_audio
from llama import detect_body_discovery_events_full_context


def process_video(video_path, json_output_dir="timestamps", audio_output_dir="audio"):
    """Process video file to extract audio, transcribe it, and detect body discovery events.

    Args:
        video_path (str): Path to the video file.
        json_output_dir (str): Directory to save the output files.
    """
    # Ensure output directory exists
    os.makedirs(json_output_dir, exist_ok=True)

    # Get the video filename without extension
    # This is used to create the audio filename
    # Also, this filename is referred to the video Id in the database
    video_filename = os.path.splitext(os.path.basename(video_path))[0]
    
    # Create audio filename in the audio output directory
    audio_filename = f"{video_filename}.wav"

    # Extract audio from video
    audio_path = os.path.join(audio_output_dir, audio_filename)
    if not os.path.exists(audio_path):
        print(f"Extracting audio from {video_path} to {audio_path}")
        # Extract audio from video
    extract_audio(video_path, audio_path)

    # Transcribe audio
    _, segments = transcribe_audio(audio_path)

    # Detect body discovery events
    structured_results = detect_body_discovery_events_full_context(segments)

    # Save structured results to JSON file with the same name as the video
    structured_results_filename = f"{video_filename}.json"
    structured_results_path = os.path.join(json_output_dir, structured_results_filename)
    with open(structured_results_path, "w") as f:
        json.dump(structured_results, f, indent=2)
    print(f"Structured results saved to {structured_results_path}")


def main():
    # Look for all video files in the download directory
    video_dir = "downloads"
    video_files = [
        f for f in os.listdir(video_dir) if f.endswith((".mp4", ".avi", ".mov"))
    ]
    print(f"Found {len(video_files)} video files in {video_dir}")

    # Process each video file
    for video_file in video_files:
        video_path = os.path.join(video_dir, video_file)
        print(f"Processing {video_path}")
        process_video(video_path)
    print("Processing complete.")


if __name__ == "__main__":
    main()
