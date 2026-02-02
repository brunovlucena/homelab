"""
Heart Rate Analyzer for Agent-Medical

Analyzes heart rate data from Apple Watch and provides clinical insights.
"""
from datetime import datetime
from typing import Optional, Dict, Any
from enum import Enum
from pydantic import BaseModel, Field


class HeartRateStatus(str, Enum):
    """Heart rate status categories."""
    NORMAL = "normal"
    TACHYCARDIA = "tachycardia"  # >100 bpm resting
    BRADYCARDIA = "bradycardia"  # <60 bpm resting
    ELEVATED = "elevated"  # Slightly high but not critical
    LOW = "low"  # Slightly low but not critical


class HeartRateContext(str, Enum):
    """Context in which heart rate was measured."""
    RESTING = "resting"
    ACTIVE = "active"
    EXERCISE = "exercise"
    SLEEP = "sleep"
    UNKNOWN = "unknown"


class HeartRateAnalysis(BaseModel):
    """Result of heart rate analysis."""
    bpm: int = Field(..., ge=0, le=250, description="Heart rate in beats per minute")
    status: HeartRateStatus
    context: HeartRateContext
    timestamp: datetime
    device: Optional[str] = None
    
    # Analysis details
    is_normal: bool = Field(default=True, description="Whether heart rate is within normal range")
    recommendation: str = Field(..., description="Clinical recommendation")
    severity: str = Field(default="none", description="Severity: none, mild, moderate, severe")
    
    # Baseline comparison (if available)
    baseline_bpm: Optional[int] = None
    deviation_from_baseline: Optional[int] = None
    
    # Reference ranges
    normal_min: int = Field(default=60, description="Minimum normal resting heart rate")
    normal_max: int = Field(default=100, description="Maximum normal resting heart rate")


class HeartRateAnalyzer:
    """Analyzes heart rate data and provides clinical insights."""
    
    # Default thresholds
    RESTING_NORMAL_MIN = 60
    RESTING_NORMAL_MAX = 100
    TACHYCARDIA_THRESHOLD = 100  # bpm
    BRADYCARDIA_THRESHOLD = 60   # bpm
    
    # Exercise thresholds (higher acceptable range)
    EXERCISE_NORMAL_MIN = 50
    EXERCISE_NORMAL_MAX = 180  # Age-dependent, simplified for now
    
    # Sleep thresholds (lower acceptable range)
    SLEEP_NORMAL_MIN = 40
    SLEEP_NORMAL_MAX = 60
    
    def __init__(
        self,
        resting_min: int = RESTING_NORMAL_MIN,
        resting_max: int = RESTING_NORMAL_MAX,
        tachycardia_threshold: int = TACHYCARDIA_THRESHOLD,
        bradycardia_threshold: int = BRADYCARDIA_THRESHOLD,
    ):
        """
        Initialize analyzer with custom thresholds.
        
        Args:
            resting_min: Minimum normal resting heart rate
            resting_max: Maximum normal resting heart rate
            tachycardia_threshold: Threshold for tachycardia (> this value)
            bradycardia_threshold: Threshold for bradycardia (< this value)
        """
        self.resting_min = resting_min
        self.resting_max = resting_max
        self.tachycardia_threshold = tachycardia_threshold
        self.bradycardia_threshold = bradycardia_threshold
    
    def analyze(
        self,
        bpm: int,
        context: HeartRateContext = HeartRateContext.RESTING,
        timestamp: Optional[datetime] = None,
        device: Optional[str] = None,
        baseline_bpm: Optional[int] = None,
    ) -> HeartRateAnalysis:
        """
        Analyze heart rate data.
        
        Args:
            bpm: Heart rate in beats per minute
            context: Context of measurement (resting, active, exercise, sleep)
            timestamp: When the measurement was taken
            device: Device that recorded the measurement
            baseline_bpm: Patient's baseline heart rate for comparison
            
        Returns:
            HeartRateAnalysis with status, recommendation, and severity
        """
        if timestamp is None:
            timestamp = datetime.utcnow()
        
        # Determine normal ranges based on context
        normal_min, normal_max = self._get_normal_range(context)
        
        # Analyze heart rate
        status, severity, recommendation = self._analyze_heart_rate(
            bpm, context, normal_min, normal_max, baseline_bpm
        )
        
        # Calculate deviation from baseline if available
        deviation_from_baseline = None
        if baseline_bpm is not None:
            deviation_from_baseline = bpm - baseline_bpm
        
        is_normal = status in [HeartRateStatus.NORMAL, HeartRateStatus.ELEVATED, HeartRateStatus.LOW]
        
        return HeartRateAnalysis(
            bpm=bpm,
            status=status,
            context=context,
            timestamp=timestamp,
            device=device,
            is_normal=is_normal,
            recommendation=recommendation,
            severity=severity,
            baseline_bpm=baseline_bpm,
            deviation_from_baseline=deviation_from_baseline,
            normal_min=normal_min,
            normal_max=normal_max,
        )
    
    def _get_normal_range(self, context: HeartRateContext) -> tuple[int, int]:
        """Get normal heart rate range for given context."""
        if context == HeartRateContext.EXERCISE:
            return (self.EXERCISE_NORMAL_MIN, self.EXERCISE_NORMAL_MAX)
        elif context == HeartRateContext.SLEEP:
            return (self.SLEEP_NORMAL_MIN, self.SLEEP_NORMAL_MAX)
        else:  # RESTING, ACTIVE, UNKNOWN
            return (self.resting_min, self.resting_max)
    
    def _analyze_heart_rate(
        self,
        bpm: int,
        context: HeartRateContext,
        normal_min: int,
        normal_max: int,
        baseline_bpm: Optional[int],
    ) -> tuple[HeartRateStatus, str, str]:
        """
        Analyze heart rate and return status, severity, and recommendation.
        
        Returns:
            Tuple of (status, severity, recommendation)
        """
        # Tachycardia (>100 bpm resting, or >180 during exercise)
        if bpm > (self.tachycardia_threshold if context == HeartRateContext.RESTING else 180):
            if context == HeartRateContext.RESTING:
                severity = "moderate" if bpm < 120 else "severe"
                recommendation = (
                    f"Elevated heart rate detected: {bpm} bpm (resting). "
                    "Consider consulting with a healthcare provider if this persists or is accompanied by symptoms."
                )
                if severity == "severe":
                    recommendation = (
                        f"Significantly elevated heart rate: {bpm} bpm (resting). "
                        "Please contact your healthcare provider or seek medical attention if experiencing symptoms."
                    )
                return (HeartRateStatus.TACHYCARDIA, severity, recommendation)
            else:
                # Exercise context - may be normal
                if bpm > 200:
                    severity = "moderate"
                    recommendation = (
                        f"Very high heart rate during {context.value}: {bpm} bpm. "
                        "Consider slowing down if experiencing discomfort."
                    )
                    return (HeartRateStatus.TACHYCARDIA, severity, recommendation)
        
        # Bradycardia (<60 bpm resting, or <40 during sleep)
        if bpm < (self.bradycardia_threshold if context != HeartRateContext.SLEEP else 40):
            if context == HeartRateContext.RESTING:
                severity = "mild" if bpm > 50 else "moderate"
                recommendation = (
                    f"Lower than normal heart rate: {bpm} bpm (resting). "
                    "This may be normal for some individuals. Monitor and consult healthcare provider if experiencing symptoms."
                )
                if severity == "moderate":
                    recommendation = (
                        f"Significantly low heart rate: {bpm} bpm (resting). "
                        "Please consult with your healthcare provider."
                    )
                return (HeartRateStatus.BRADYCARDIA, severity, recommendation)
            elif context == HeartRateContext.SLEEP:
                # Lower HR during sleep is normal
                return (HeartRateStatus.NORMAL, "none", "Heart rate is within normal range for sleep.")
        
        # Check if slightly elevated or low (within acceptable range but noteworthy)
        if context == HeartRateContext.RESTING:
            if bpm > normal_max - 10:  # Close to upper limit
                return (
                    HeartRateStatus.ELEVATED,
                    "none",
                    f"Heart rate is slightly elevated: {bpm} bpm. Continue monitoring."
                )
            elif bpm < normal_min + 10:  # Close to lower limit
                return (
                    HeartRateStatus.LOW,
                    "none",
                    f"Heart rate is slightly low: {bpm} bpm. Continue monitoring."
                )
        
        # Normal range
        return (
            HeartRateStatus.NORMAL,
            "none",
            f"Heart rate is within normal range: {bpm} bpm ({context.value})."
        )


# Global analyzer instance
_analyzer: Optional[HeartRateAnalyzer] = None


def get_analyzer() -> HeartRateAnalyzer:
    """Get or create global analyzer instance."""
    global _analyzer
    if _analyzer is None:
        # Can be configured via environment variables
        import os
        resting_min = int(os.getenv("HEART_RATE_NORMAL_MIN", "60"))
        resting_max = int(os.getenv("HEART_RATE_NORMAL_MAX", "100"))
        tachycardia_threshold = int(os.getenv("HEART_RATE_TACHYCARDIA_THRESHOLD", "100"))
        bradycardia_threshold = int(os.getenv("HEART_RATE_BRADYCARDIA_THRESHOLD", "60"))
        
        _analyzer = HeartRateAnalyzer(
            resting_min=resting_min,
            resting_max=resting_max,
            tachycardia_threshold=tachycardia_threshold,
            bradycardia_threshold=bradycardia_threshold,
        )
    return _analyzer


