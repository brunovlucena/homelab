"""
PDF Extraction Service using LangExtract

Extracts structured medical data from PDF exam documents using langextract.
Uses Ollama for local model inference (no API key required).
"""
import os
import io
import asyncio
from typing import Optional, Dict, Any, List
import structlog

try:
    import langextract as lx
    import pdfplumber
    LANGEXTRACT_AVAILABLE = True
except ImportError:
    LANGEXTRACT_AVAILABLE = False
    lx = None
    pdfplumber = None

logger = structlog.get_logger()


# Define extraction prompt and examples for medical exams
MEDICAL_EXAM_PROMPT = """
Extract structured medical information from the exam document.

Extract the following information:
- Patient information (name, date of birth, patient ID)
- Exam type and date
- Test results with values, units, and reference ranges
- Findings and interpretations
- Physician notes and recommendations
- Medications mentioned (name, dosage, frequency)
- Diagnoses and conditions
- Vital signs and measurements

Use exact text from the document. Do not paraphrase. Provide meaningful attributes for context.
"""

def get_medical_exam_examples():
    """Get medical exam extraction examples for LangExtract."""
    if not LANGEXTRACT_AVAILABLE:
        return []
    
    return [
        lx.data.ExampleData(
            text="""
            MEDICAL EXAM REPORT
            Patient: John Doe
            DOB: 1990-01-15
            Exam Date: 2024-01-10
            
            Complete Blood Count (CBC):
            - Hemoglobin: 14.5 g/dL (Reference: 12.0-16.0)
            - Hematocrit: 42.0% (Reference: 36-48)
            - White Blood Cell Count: 7.2 x10^3/μL (Reference: 4.0-11.0)
            
            Findings: All values within normal range.
            """,
            extractions=[
                lx.data.Extraction(
                    extraction_class="patient_info",
                    extraction_text="John Doe",
                    attributes={"dob": "1990-01-15", "type": "name"}
                ),
                lx.data.Extraction(
                    extraction_class="exam_info",
                    extraction_text="Complete Blood Count (CBC)",
                    attributes={"date": "2024-01-10", "type": "lab_test"}
                ),
                lx.data.Extraction(
                    extraction_class="test_result",
                    extraction_text="Hemoglobin: 14.5 g/dL",
                    attributes={
                        "test_name": "Hemoglobin",
                        "value": "14.5",
                        "unit": "g/dL",
                        "reference_range": "12.0-16.0",
                        "status": "normal"
                    }
                ),
                lx.data.Extraction(
                    extraction_class="finding",
                    extraction_text="All values within normal range",
                    attributes={"type": "interpretation", "status": "normal"}
                ),
            ]
        )
    ]


class PDFExtractor:
    """Service for extracting structured data from medical exam PDFs."""
    
    def __init__(
        self,
        model_id: str = "llama3.2:3b",
        model_url: Optional[str] = None,
        use_ollama: bool = True
    ):
        """
        Initialize PDF extractor.
        
        Args:
            model_id: Ollama model ID (default: llama3.2:3b)
            model_url: Ollama URL (defaults to OLLAMA_URL env var or localhost:11434)
            use_ollama: Whether to use Ollama (default: True)
        """
        if not LANGEXTRACT_AVAILABLE:
            raise ImportError(
                "langextract is not installed. Install it with: pip install langextract"
            )
        
        self.model_id = model_id
        self.use_ollama = use_ollama
        self.model_url = model_url or os.getenv("OLLAMA_URL", "http://localhost:11434")
        
        logger.info(
            "pdf_extractor_initialized",
            model_id=self.model_id,
            model_url=self.model_url,
            use_ollama=self.use_ollama
        )
    
    def extract_text_from_pdf(self, pdf_bytes: bytes) -> str:
        """
        Extract text content from PDF bytes.
        
        Args:
            pdf_bytes: PDF file content as bytes
            
        Returns:
            Extracted text from PDF
        """
        try:
            # Use pdfplumber for better text extraction
            with pdfplumber.open(io.BytesIO(pdf_bytes)) as pdf:
                text_parts = []
                for page in pdf.pages:
                    text = page.extract_text()
                    if text:
                        text_parts.append(text)
                return "\n\n".join(text_parts)
        except Exception as e:
            logger.error("pdf_text_extraction_failed", error=str(e))
            raise ValueError(f"Failed to extract text from PDF: {str(e)}")
    
    async def extract_medical_data(
        self,
        pdf_bytes: bytes,
        patient_id: Optional[str] = None,
        metadata: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        Extract structured medical data from PDF using LangExtract.
        
        Args:
            pdf_bytes: PDF file content as bytes
            patient_id: Optional patient ID to associate with extraction
            metadata: Optional metadata to include in extraction
            
        Returns:
            Dictionary with extracted data and metadata
        """
        # Extract text from PDF
        pdf_text = self.extract_text_from_pdf(pdf_bytes)
        
        if not pdf_text or not pdf_text.strip():
            raise ValueError("PDF appears to be empty or could not extract text")
        
        logger.info(
            "extracting_medical_data",
            text_length=len(pdf_text),
            patient_id=patient_id,
            model=self.model_id
        )
        
        try:
            # Run LangExtract extraction with Ollama
            examples = get_medical_exam_examples()
            
            # For Ollama models, use model_url and disable schema constraints
            extract_params = {
                "text_or_documents": pdf_text,
                "prompt_description": MEDICAL_EXAM_PROMPT,
                "examples": examples,
                "model_id": self.model_id,
                "extraction_passes": 2,  # Multiple passes for better recall
                "max_workers": 4,  # Parallel processing
                "max_char_buffer": 2000,  # Smaller contexts for better accuracy
                "fence_output": False,  # Required for Ollama
                "use_schema_constraints": False,  # Required for Ollama
            }
            
            # Add Ollama-specific parameters
            if self.use_ollama:
                extract_params["model_url"] = self.model_url
            
            # Run extraction in executor since lx.extract is synchronous
            result = await asyncio.to_thread(lx.extract, **extract_params)
            
            # Process extractions into structured format
            extractions = []
            if hasattr(result, 'documents') and result.documents:
                # Handle multiple documents (if document was chunked)
                for doc in result.documents:
                    if hasattr(doc, 'extractions') and doc.extractions:
                        for ext in doc.extractions:
                            extractions.append({
                                "class": ext.extraction_class if hasattr(ext, 'extraction_class') else "unknown",
                                "text": ext.extraction_text if hasattr(ext, 'extraction_text') else str(ext),
                                "attributes": ext.attributes if hasattr(ext, 'attributes') else {},
                            })
            elif hasattr(result, 'extractions') and result.extractions:
                # Handle single document extractions
                for ext in result.extractions:
                    extractions.append({
                        "class": ext.extraction_class if hasattr(ext, 'extraction_class') else "unknown",
                        "text": ext.extraction_text if hasattr(ext, 'extraction_text') else str(ext),
                        "attributes": ext.attributes if hasattr(ext, 'attributes') else {},
                    })
            
            # Group extractions by class for easier querying
            grouped_extractions = {}
            for ext in extractions:
                ext_class = ext["class"]
                if ext_class not in grouped_extractions:
                    grouped_extractions[ext_class] = []
                grouped_extractions[ext_class].append(ext)
            
            extraction_result = {
                "raw_text": pdf_text,
                "extractions": extractions,
                "grouped_extractions": grouped_extractions,
                "patient_id": patient_id,
                "metadata": metadata or {},
                "model_used": self.model_id,
                "model_url": self.model_url if self.use_ollama else None,
                "extraction_count": len(extractions),
            }
            
            logger.info(
                "extraction_completed",
                extraction_count=len(extractions),
                classes=list(grouped_extractions.keys()),
                patient_id=patient_id
            )
            
            return extraction_result
            
        except Exception as e:
            logger.error("langextract_extraction_failed", error=str(e), patient_id=patient_id)
            raise ValueError(f"Failed to extract medical data: {str(e)}")


# Global extractor instance (initialized on first use)
_extractor_instance: Optional[PDFExtractor] = None


def get_pdf_extractor(
    model_id: Optional[str] = None,
    model_url: Optional[str] = None,
    use_ollama: bool = True
) -> PDFExtractor:
    """
    Get or create PDF extractor instance.
    
    Args:
        model_id: Ollama model ID (defaults to OLLAMA_MODEL env var or llama3.2:3b)
        model_url: Ollama URL (defaults to OLLAMA_URL env var)
        use_ollama: Whether to use Ollama (default: True)
    """
    global _extractor_instance
    
    # Get defaults from environment or use sensible defaults
    default_model_id = os.getenv("OLLAMA_MODEL", "llama3.2:3b")
    model_id = model_id or default_model_id
    
    if _extractor_instance is None or _extractor_instance.model_id != model_id:
        _extractor_instance = PDFExtractor(
            model_id=model_id,
            model_url=model_url,
            use_ollama=use_ollama
        )
    return _extractor_instance
