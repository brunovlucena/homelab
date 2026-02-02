"""
OCR - Text Extraction from Screenshots

Extrai texto de imagens usando EasyOCR ou Tesseract.
"""

import os
import logging
from typing import Optional, List, Dict
import base64
import io

logger = logging.getLogger(__name__)

# Try to import OCR libraries
try:
    import easyocr
    EASYOCR_AVAILABLE = True
except ImportError:
    EASYOCR_AVAILABLE = False
    easyocr = None

try:
    import pytesseract
    from PIL import Image
    TESSERACT_AVAILABLE = True
except ImportError:
    TESSERACT_AVAILABLE = False
    pytesseract = None
    Image = None


class OCRProcessor:
    """Processa OCR em screenshots."""
    
    def __init__(self, prefer_easyocr: bool = True):
        """
        Inicializa o processador OCR.
        
        Args:
            prefer_easyocr: Se True, usa EasyOCR (mais preciso, mais lento).
                           Se False, usa Tesseract (mais rápido).
        """
        self.prefer_easyocr = prefer_easyocr
        self.easyocr_reader = None
        
        # Initialize EasyOCR if available and preferred
        if prefer_easyocr and EASYOCR_AVAILABLE:
            try:
                # Use English and Portuguese for better coverage
                self.easyocr_reader = easyocr.Reader(['en', 'pt'], gpu=False)
                logger.info("EasyOCR initialized successfully")
            except Exception as e:
                logger.warning(f"Failed to initialize EasyOCR: {e}")
                self.easyocr_reader = None
        
        # Check Tesseract availability
        if not TESSERACT_AVAILABLE and not self.easyocr_reader:
            logger.warning("No OCR library available. Install easyocr or pytesseract.")
    
    def extract_text_from_image(self, image_bytes: bytes) -> Dict[str, any]:
        """
        Extrai texto de uma imagem.
        
        Args:
            image_bytes: Bytes da imagem (PNG, JPEG, etc.)
        
        Returns:
            Dicionário com texto extraído e metadados
        """
        if not image_bytes:
            return {"text": "", "confidence": 0.0, "method": "none"}
        
        # Try EasyOCR first if available
        if self.prefer_easyocr and self.easyocr_reader:
            return self._extract_with_easyocr(image_bytes)
        
        # Fallback to Tesseract
        if TESSERACT_AVAILABLE:
            return self._extract_with_tesseract(image_bytes)
        
        # No OCR available
        logger.warning("No OCR library available")
        return {"text": "", "confidence": 0.0, "method": "none"}
    
    def _extract_with_easyocr(self, image_bytes: bytes) -> Dict[str, any]:
        """Extrai texto usando EasyOCR."""
        try:
            # EasyOCR works with image arrays
            from PIL import Image
            import numpy as np
            
            image = Image.open(io.BytesIO(image_bytes))
            image_array = np.array(image)
            
            # Extract text
            results = self.easyocr_reader.readtext(image_array)
            
            # Combine all text
            text_parts = []
            confidences = []
            
            for (bbox, text, confidence) in results:
                text_parts.append(text)
                confidences.append(confidence)
            
            full_text = "\n".join(text_parts)
            avg_confidence = sum(confidences) / len(confidences) if confidences else 0.0
            
            logger.info(f"EasyOCR extracted {len(text_parts)} text blocks, avg confidence: {avg_confidence:.2f}")
            
            return {
                "text": full_text,
                "confidence": avg_confidence,
                "method": "easyocr",
                "blocks": [
                    {"text": text, "confidence": conf}
                    for (_, text, conf) in results
                ],
            }
            
        except Exception as e:
            logger.error(f"EasyOCR extraction failed: {e}")
            # Fallback to Tesseract
            if TESSERACT_AVAILABLE:
                return self._extract_with_tesseract(image_bytes)
            return {"text": "", "confidence": 0.0, "method": "easyocr_error", "error": str(e)}
    
    def _extract_with_tesseract(self, image_bytes: bytes) -> Dict[str, any]:
        """Extrai texto usando Tesseract."""
        try:
            image = Image.open(io.BytesIO(image_bytes))
            
            # Extract text
            text = pytesseract.image_to_string(image, lang='eng+por')
            
            # Get confidence data
            data = pytesseract.image_to_data(image, output_type=pytesseract.Output.DICT)
            confidences = [conf for conf in data['conf'] if conf > 0]
            avg_confidence = sum(confidences) / len(confidences) if confidences else 0.0
            avg_confidence = avg_confidence / 100.0  # Normalize to 0-1
            
            logger.info(f"Tesseract extracted text, avg confidence: {avg_confidence:.2f}")
            
            return {
                "text": text.strip(),
                "confidence": avg_confidence,
                "method": "tesseract",
            }
            
        except Exception as e:
            logger.error(f"Tesseract extraction failed: {e}")
            return {"text": "", "confidence": 0.0, "method": "tesseract_error", "error": str(e)}


# Global instance (lazy initialization)
_ocr_processor: Optional[OCRProcessor] = None


def get_ocr_processor() -> OCRProcessor:
    """Get or create global OCR processor instance."""
    global _ocr_processor
    if _ocr_processor is None:
        _ocr_processor = OCRProcessor()
    return _ocr_processor
