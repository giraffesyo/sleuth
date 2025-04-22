import whisper
import ffmpeg


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
        return audio_path
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
