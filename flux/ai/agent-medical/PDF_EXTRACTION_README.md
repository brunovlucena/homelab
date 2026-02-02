# PDF Medical Exam Extraction with LangExtract

This document describes the integration of LangExtract for extracting structured data from medical exam PDFs.

## Overview

The medical agent now supports:
1. **PDF Upload**: Users can upload PDF medical exam documents
2. **Automatic Extraction**: LangExtract automatically extracts structured medical data from PDFs
3. **Storage**: PDFs and extracted data are persisted in MongoDB for future fine-tuning
4. **iOS Integration**: The iOS app includes PDF upload functionality

## Architecture

### Backend (agent-medical)

#### Components

1. **PDF Extractor Service** (`src/medical_agent/pdf_extractor.py`)
   - Extracts text from PDFs using `pdfplumber`
   - Uses LangExtract to extract structured medical data
   - Supports multiple extraction passes for better recall
   - Configurable model (default: `gemini-2.5-flash`)

2. **Database Schema** (`src/shared/database.py`)
   - New collection: `medical_exam_pdfs`
   - Stores PDF metadata, extracted data, and storage paths
   - Methods:
     - `store_medical_exam_pdf()`: Store PDF and extracted data
     - `get_medical_exam_pdfs()`: Retrieve PDFs for a patient
     - `get_medical_exam_pdf()`: Get a specific PDF by ID

3. **API Endpoints** (`src/medical_agent/main.py`)
   - `POST /api/v1/exams/upload`: Upload PDF and extract data
   - `GET /api/v1/exams?patient_id=...`: List PDFs for a patient

#### Local Model Support (Ollama)

The extractor uses Ollama for local inference, which means:
- **No API keys required**: Runs entirely on your infrastructure
- **Privacy**: All data stays local
- **Cost-effective**: No cloud API costs
- **Configurable**: Use any Ollama model (llama3.2, mistral, etc.)

The extractor automatically detects Ollama and configures LangExtract accordingly.

#### Extraction Prompt

The extraction is configured with a medical exam-specific prompt that extracts:
- Patient information (name, DOB, patient ID)
- Exam type and date
- Test results with values, units, and reference ranges
- Findings and interpretations
- Physician notes and recommendations
- Medications mentioned
- Diagnoses and conditions
- Vital signs and measurements

### Frontend (AppMedical iOS App)

#### Components

1. **PDFUploadService** (`AppMedical/Services/PDFUploadService.swift`)
   - Handles multipart form data upload
   - Manages authentication
   - Parses extraction response

2. **DocumentPicker** (`AppMedical/Views/DocumentPicker.swift`)
   - SwiftUI wrapper for `UIDocumentPickerViewController`
   - Allows users to select PDF files

3. **ChatViewModel** (`AppMedical/ViewModels/ChatViewModel.swift`)
   - Added `uploadPDF()` method
   - Integrates with PDFUploadService
   - Shows upload progress and results

4. **ContentView** (`AppMedical/Views/ContentView.swift`)
   - Added PDF attachment button in chat input
   - Opens document picker on tap

## Usage

### Setting up Ollama (Local Models)

The PDF extractor uses Ollama for local model inference. No API key is required.

1. **Install Ollama**: Download from [ollama.com](https://ollama.com)

2. **Pull a model** (if not already available):
   ```bash
   ollama pull llama3.2:3b
   # Or use a larger model for better extraction quality:
   ollama pull llama3.2:1b
   ollama pull mistral
   ```

3. **Ensure Ollama is running**:
   ```bash
   ollama serve
   ```

4. **Configure environment variables** (optional, defaults provided):
   ```bash
   export OLLAMA_URL="http://localhost:11434"
   export OLLAMA_MODEL="llama3.2:3b"
   ```

### Uploading a PDF via API

```bash
curl -X POST http://localhost:8080/api/v1/exams/upload \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@exam.pdf" \
  -F "patient_id=patient-123" \
  -F 'metadata={"source": "api", "exam_type": "lab_results"}'
```

### Uploading from iOS App

1. Open the medical agent chat
2. Tap the document icon (📄) in the input area
3. Select a PDF from your device
4. The app uploads the PDF and shows extraction results

## Data Storage

### MongoDB Collection: `medical_exam_pdfs`

Document structure:
```javascript
{
  "_id": ObjectId("..."),
  "patient_id": "patient-123",
  "uploaded_by": "doctor-456",
  "uploaded_by_role": "doctor",
  "filename": "lab_results_2024.pdf",
  "storage_path": "medical_exams/patient-123/uuid/filename.pdf",
  "file_size": 123456,
  "extracted_data": {
    "raw_text": "...",
    "extractions": [...],
    "grouped_extractions": {
      "patient_info": [...],
      "test_result": [...],
      "finding": [...]
    },
    "patient_id": "patient-123",
    "metadata": {...},
    "model_used": "gemini-2.5-flash",
    "extraction_count": 42
  },
  "metadata": {...},
  "created_at": ISODate("2024-01-15T10:30:00Z"),
  "updated_at": ISODate("2024-01-15T10:30:00Z")
}
```

### Indexes

Recommended indexes for performance:
```javascript
db.medical_exam_pdfs.createIndex({ "patient_id": 1, "created_at": -1 })
db.medical_exam_pdfs.createIndex({ "uploaded_by": 1 })
```

## Future Fine-Tuning

The extracted data structure is designed to support future fine-tuning:

1. **Training Data Format**: The `extractions` array contains structured examples that can be used for fine-tuning
2. **Structured Output**: Extractions are grouped by class for easy querying and analysis
3. **Metadata Preservation**: All original text and metadata are preserved

### Fine-Tuning Workflow (Planned)

1. Collect uploaded PDFs and their extractions
2. Review and validate extractions (human-in-the-loop)
3. Create training dataset from validated extractions
4. Fine-tune model on domain-specific medical data
5. Deploy fine-tuned model

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `OLLAMA_URL` | Ollama server URL | `http://localhost:11434` |
| `OLLAMA_MODEL` | Ollama model ID | `llama3.2:3b` |

### Model Selection

Recommended Ollama models for medical extraction:
- **llama3.2:3b**: Fast, good balance of speed and quality (default)
- **llama3.2:1b**: Faster, lower quality
- **mistral**: Good quality, medium speed
- **llama3.1:8b**: Higher quality, slower

To use a different model:
```bash
export OLLAMA_MODEL="mistral"
# Or pull it first:
ollama pull mistral
```

## Security & Compliance

- **RBAC**: PDF uploads respect role-based access control
- **HIPAA Compliance**: All uploads are logged in audit trail
- **Patient Data Isolation**: Users can only upload/view PDFs for patients they have access to
- **Data Encryption**: PDFs should be stored encrypted (MinIO encryption recommended for production)

## Production Considerations

### Storage Optimization

Currently, PDF bytes are stored in MongoDB. For production:

1. **Use MinIO for PDF Storage**: Store PDFs in MinIO and only store references in MongoDB
2. **Implement Cleanup**: Archive or delete old PDFs based on retention policy
3. **Compression**: Compress PDFs before storage

### Example MinIO Integration (Future)

```python
# Upload PDF to MinIO
minio_client.put_object(
    bucket="medical-exams",
    object_name=storage_path,
    data=pdf_bytes,
    length=len(pdf_bytes),
    content_type="application/pdf"
)

# Store only reference in MongoDB
doc = {
    "patient_id": patient_id,
    "storage_path": storage_path,
    "minio_bucket": "medical-exams",
    # ... rest of document
}
```

## Testing

### Test PDF Upload

```python
import requests

url = "http://localhost:8080/api/v1/exams/upload"
headers = {"Authorization": "Bearer YOUR_TOKEN"}

with open("test_exam.pdf", "rb") as f:
    files = {"file": ("test_exam.pdf", f, "application/pdf")}
    data = {
        "patient_id": "patient-123",
        "metadata": '{"source": "test"}'
    }
    response = requests.post(url, headers=headers, files=files, data=data)
    print(response.json())
```

## Dependencies

New dependencies added to `requirements.txt`:
- `langextract==1.1.1`: PDF extraction library
- `pypdf2==3.0.1`: PDF parsing
- `pdfplumber==0.10.3`: Better text extraction
- `PyPDFium2==4.27.0`: PDF processing

## References

- [LangExtract Documentation](https://pypi.org/project/langextract/)
- [LangExtract GitHub](https://github.com/google/langextract)
- Medical Agent Architecture: See `README.md` in this directory
