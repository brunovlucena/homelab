"""
Audio Analyzer Service for YouTube Ad Detection

This service provides:
1. Audio fingerprinting using Chromaprint
2. Ad signature database management
3. Real-time audio matching for ad detection

The service maintains a database of known ad audio fingerprints
and can match incoming audio against this database.
"""
import os
import hashlib
import subprocess
import tempfile
from typing import Optional, List, Dict
from datetime import datetime
import json

from fastapi import FastAPI, HTTPException, UploadFile, File, BackgroundTasks
from pydantic import BaseModel
import redis
import httpx

app = FastAPI(
    title="Audio Analyzer Service",
    description="Audio fingerprinting for YouTube ad detection",
    version="1.0.0"
)

# Configuration
REDIS_URL = os.environ.get("REDIS_URL", "redis://localhost:6379")
FINGERPRINT_DURATION = 30  # Analyze first 30 seconds
MATCH_THRESHOLD = 0.7  # 70% similarity threshold

# Initialize Redis connection
redis_client = None
try:
    redis_client = redis.from_url(REDIS_URL)
    redis_client.ping()
except Exception as e:
    print(f"Redis not available: {e}")


class AudioFingerprint(BaseModel):
    """Audio fingerprint data model."""
    fingerprint: str
    duration: float
    source: str  # 'ad' or 'content'
    metadata: Dict = {}


class MatchResult(BaseModel):
    """Result of fingerprint matching."""
    is_ad: bool
    confidence: float
    matched_fingerprint: Optional[str] = None
    matched_source: Optional[str] = None


class AdSignature(BaseModel):
    """Known ad signature for database."""
    fingerprint: str
    ad_name: str
    ad_duration: float
    added_at: str
    metadata: Dict = {}


# In-memory cache for fingerprints (fallback if Redis unavailable)
local_ad_db: Dict[str, AdSignature] = {}


def get_chromaprint(audio_data: bytes) -> Optional[str]:
    """
    Generate Chromaprint fingerprint from audio data.
    
    Uses fpcalc (Chromaprint CLI) for fingerprint generation.
    """
    try:
        # Write audio to temp file
        with tempfile.NamedTemporaryFile(suffix='.mp3', delete=False) as f:
            f.write(audio_data)
            temp_path = f.name
        
        # Run fpcalc
        result = subprocess.run(
            ['fpcalc', '-raw', '-length', str(FINGERPRINT_DURATION), temp_path],
            capture_output=True,
            text=True,
            timeout=30
        )
        
        # Clean up
        os.unlink(temp_path)
        
        if result.returncode == 0:
            # Parse fingerprint from output
            for line in result.stdout.split('\n'):
                if line.startswith('FINGERPRINT='):
                    return line.split('=', 1)[1]
        
        return None
        
    except Exception as e:
        print(f"Chromaprint error: {e}")
        return None


def extract_audio_from_video(video_data: bytes) -> Optional[bytes]:
    """Extract audio track from video data using ffmpeg."""
    try:
        with tempfile.NamedTemporaryFile(suffix='.mp4', delete=False) as video_file:
            video_file.write(video_data)
            video_path = video_file.name
        
        audio_path = video_path.replace('.mp4', '.mp3')
        
        # Extract audio using ffmpeg
        result = subprocess.run([
            'ffmpeg', '-i', video_path,
            '-vn',  # No video
            '-acodec', 'libmp3lame',
            '-ar', '22050',  # Sample rate
            '-ac', '1',  # Mono
            '-t', str(FINGERPRINT_DURATION),  # Duration limit
            '-y',  # Overwrite
            audio_path
        ], capture_output=True, timeout=60)
        
        # Clean up video file
        os.unlink(video_path)
        
        if result.returncode == 0 and os.path.exists(audio_path):
            with open(audio_path, 'rb') as f:
                audio_data = f.read()
            os.unlink(audio_path)
            return audio_data
        
        return None
        
    except Exception as e:
        print(f"Audio extraction error: {e}")
        return None


def compare_fingerprints(fp1: str, fp2: str) -> float:
    """
    Compare two fingerprints and return similarity score (0-1).
    
    Uses a simple comparison based on matching segments.
    For production, use proper fingerprint comparison algorithms.
    """
    try:
        # Convert fingerprint strings to integer arrays
        arr1 = [int(x) for x in fp1.split(',') if x]
        arr2 = [int(x) for x in fp2.split(',') if x]
        
        if not arr1 or not arr2:
            return 0.0
        
        # Calculate Hamming distance for overlapping portions
        min_len = min(len(arr1), len(arr2))
        matches = 0
        total_bits = 0
        
        for i in range(min_len):
            # XOR and count differing bits
            xor = arr1[i] ^ arr2[i]
            diff_bits = bin(xor).count('1')
            total_bits += 32  # Assuming 32-bit integers
            matches += 32 - diff_bits
        
        return matches / total_bits if total_bits > 0 else 0.0
        
    except Exception as e:
        print(f"Fingerprint comparison error: {e}")
        return 0.0


def store_fingerprint(key: str, signature: AdSignature):
    """Store fingerprint in Redis or local cache."""
    data = signature.model_dump_json()
    if redis_client:
        try:
            redis_client.hset("ad_fingerprints", key, data)
            return
        except:
            pass
    local_ad_db[key] = signature


def get_all_fingerprints() -> Dict[str, AdSignature]:
    """Retrieve all stored fingerprints."""
    if redis_client:
        try:
            data = redis_client.hgetall("ad_fingerprints")
            return {
                k.decode(): AdSignature(**json.loads(v.decode()))
                for k, v in data.items()
            }
        except:
            pass
    return local_ad_db


# API Endpoints

@app.get("/health")
async def health_check():
    """Health check endpoint."""
    return {
        "status": "healthy",
        "redis_connected": redis_client is not None,
        "fingerprints_stored": len(get_all_fingerprints())
    }


@app.post("/fingerprint", response_model=AudioFingerprint)
async def generate_fingerprint(file: UploadFile = File(...)):
    """
    Generate audio fingerprint from uploaded audio/video file.
    """
    content = await file.read()
    
    # Check if it's video and extract audio
    if file.filename and any(file.filename.endswith(ext) for ext in ['.mp4', '.webm', '.mkv']):
        audio_data = extract_audio_from_video(content)
        if not audio_data:
            raise HTTPException(status_code=400, detail="Failed to extract audio from video")
    else:
        audio_data = content
    
    fingerprint = get_chromaprint(audio_data)
    if not fingerprint:
        raise HTTPException(status_code=400, detail="Failed to generate fingerprint")
    
    return AudioFingerprint(
        fingerprint=fingerprint,
        duration=FINGERPRINT_DURATION,
        source="unknown",
        metadata={"filename": file.filename}
    )


@app.post("/match", response_model=MatchResult)
async def match_fingerprint(fingerprint: str):
    """
    Match a fingerprint against known ad signatures.
    """
    ad_db = get_all_fingerprints()
    
    best_match = None
    best_score = 0.0
    
    for key, signature in ad_db.items():
        score = compare_fingerprints(fingerprint, signature.fingerprint)
        if score > best_score:
            best_score = score
            best_match = signature
    
    is_ad = best_score >= MATCH_THRESHOLD
    
    return MatchResult(
        is_ad=is_ad,
        confidence=best_score,
        matched_fingerprint=best_match.fingerprint if best_match else None,
        matched_source=best_match.ad_name if best_match else None
    )


@app.post("/match/audio", response_model=MatchResult)
async def match_audio_file(file: UploadFile = File(...)):
    """
    Upload audio/video file and check if it matches known ads.
    """
    content = await file.read()
    
    # Extract audio if video
    if file.filename and any(file.filename.endswith(ext) for ext in ['.mp4', '.webm', '.mkv']):
        audio_data = extract_audio_from_video(content)
        if not audio_data:
            raise HTTPException(status_code=400, detail="Failed to extract audio")
    else:
        audio_data = content
    
    fingerprint = get_chromaprint(audio_data)
    if not fingerprint:
        raise HTTPException(status_code=400, detail="Failed to generate fingerprint")
    
    # Match against database
    return await match_fingerprint(fingerprint)


@app.post("/ads/register")
async def register_ad_signature(
    ad_name: str,
    file: UploadFile = File(...)
):
    """
    Register a known ad signature in the database.
    """
    content = await file.read()
    
    # Extract audio if video
    if file.filename and any(file.filename.endswith(ext) for ext in ['.mp4', '.webm', '.mkv']):
        audio_data = extract_audio_from_video(content)
        if not audio_data:
            raise HTTPException(status_code=400, detail="Failed to extract audio")
    else:
        audio_data = content
    
    fingerprint = get_chromaprint(audio_data)
    if not fingerprint:
        raise HTTPException(status_code=400, detail="Failed to generate fingerprint")
    
    # Create signature
    signature = AdSignature(
        fingerprint=fingerprint,
        ad_name=ad_name,
        ad_duration=FINGERPRINT_DURATION,
        added_at=datetime.utcnow().isoformat(),
        metadata={"filename": file.filename}
    )
    
    # Store with hash as key
    key = hashlib.sha256(fingerprint.encode()).hexdigest()[:16]
    store_fingerprint(key, signature)
    
    return {
        "status": "registered",
        "key": key,
        "ad_name": ad_name
    }


@app.get("/ads")
async def list_ad_signatures():
    """List all registered ad signatures."""
    ad_db = get_all_fingerprints()
    return {
        "count": len(ad_db),
        "signatures": [
            {
                "key": k,
                "ad_name": v.ad_name,
                "ad_duration": v.ad_duration,
                "added_at": v.added_at
            }
            for k, v in ad_db.items()
        ]
    }


@app.delete("/ads/{key}")
async def delete_ad_signature(key: str):
    """Delete an ad signature from the database."""
    if redis_client:
        try:
            redis_client.hdel("ad_fingerprints", key)
        except:
            pass
    
    if key in local_ad_db:
        del local_ad_db[key]
    
    return {"status": "deleted", "key": key}


@app.post("/analyze/url")
async def analyze_from_url(url: str, background_tasks: BackgroundTasks):
    """
    Analyze audio from a URL (e.g., YouTube video segment).
    Downloads the content and generates fingerprint.
    """
    async def download_and_analyze():
        try:
            async with httpx.AsyncClient() as client:
                response = await client.get(url, timeout=60)
                if response.status_code == 200:
                    audio_data = extract_audio_from_video(response.content)
                    if audio_data:
                        fingerprint = get_chromaprint(audio_data)
                        # Store result in Redis for later retrieval
                        if fingerprint and redis_client:
                            redis_client.setex(
                                f"analysis:{hashlib.sha256(url.encode()).hexdigest()[:16]}",
                                3600,
                                fingerprint
                            )
        except Exception as e:
            print(f"URL analysis error: {e}")
    
    background_tasks.add_task(download_and_analyze)
    
    return {
        "status": "processing",
        "message": "Analysis started in background"
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)

