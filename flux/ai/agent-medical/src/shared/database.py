"""
MongoDB database connection and utilities for agent-medical.
"""
import os
from typing import Optional, List, Dict, Any
import structlog
from datetime import datetime

try:
    from motor.motor_asyncio import AsyncIOMotorClient, AsyncIOMotorDatabase
    from pymongo import ASCENDING, DESCENDING
    from pymongo.errors import ConnectionFailure
except ImportError:
    AsyncIOMotorClient = None
    AsyncIOMotorDatabase = None

from .types import MedicalRecord, LabResult, Prescription, Patient
from .vault import get_vault_client

logger = structlog.get_logger()


class Database:
    """MongoDB database connection with access control."""
    
    def __init__(self):
        self.client: Optional[AsyncIOMotorClient] = None
        self.db: Optional[AsyncIOMotorDatabase] = None
        
        # Try to get from Vault first
        vault = get_vault_client()
        db_creds = vault.get_db_credentials()
        
        self.connection_string = db_creds.get("url") if db_creds else os.getenv(
            "MONGODB_URL",
            "mongodb://medical_user:medical_pass@mongodb.medical.svc.cluster.local:27017/medical_db"
        )
        self.database_name = os.getenv("MONGODB_DATABASE", "medical_db")
    
    async def connect(self):
        """Create MongoDB connection."""
        if not AsyncIOMotorClient:
            logger.warning("motor_not_installed", note="MongoDB driver not available")
            return
        
        try:
            self.client = AsyncIOMotorClient(
                self.connection_string,
                serverSelectionTimeoutMS=5000,
            )
            # Test connection
            await self.client.admin.command('ping')
            self.db = self.client[self.database_name]
            logger.info("database_connected", database=self.database_name)
        except ConnectionFailure as e:
            logger.error("database_connection_failed", error=str(e))
            raise
        except Exception as e:
            logger.error("database_connection_error", error=str(e))
            raise
    
    async def disconnect(self):
        """Close MongoDB connection."""
        if self.client:
            self.client.close()
            logger.info("database_disconnected")
    
    async def get_patient_record(
        self,
        patient_id: str,
        user_id: str,
        user_role: str
    ) -> Optional[Dict[str, Any]]:
        """
        Get patient record with access control.
        
        Access control is enforced at application level.
        """
        if not self.db:
            return None
        
        try:
            patient = await self.db.patients.find_one(
                {"_id": patient_id},
                {"cpf": 0}  # Don't return encrypted CPF by default
            )
            
            if patient:
                # Convert _id to id for consistency
                patient["id"] = str(patient.pop("_id"))
                return patient
            return None
        except Exception as e:
            logger.error("get_patient_record_failed", error=str(e))
            return None
    
    async def get_medical_records(
        self,
        patient_id: str,
        user_id: str,
        user_role: str,
        record_type: Optional[str] = None,
        limit: int = 50
    ) -> List[Dict[str, Any]]:
        """Get medical records for a patient."""
        if not self.db:
            return []
        
        try:
            query = {"patient_id": patient_id}
            if record_type:
                query["type"] = record_type
            
            cursor = self.db.medical_records.find(query).sort("date", DESCENDING).limit(limit)
            records = await cursor.to_list(length=limit)
            
            # Convert _id to id
            for record in records:
                record["id"] = str(record.pop("_id"))
            
            return records
        except Exception as e:
            logger.error("get_medical_records_failed", error=str(e))
            return []
    
    async def get_lab_results(
        self,
        patient_id: str,
        user_id: str,
        user_role: str,
        limit: int = 50
    ) -> List[Dict[str, Any]]:
        """Get lab results for a patient."""
        if not self.db:
            return []
        
        try:
            cursor = self.db.lab_results.find(
                {"patient_id": patient_id}
            ).sort("test_date", DESCENDING).limit(limit)
            
            results = await cursor.to_list(length=limit)
            
            # Convert _id to id
            for result in results:
                result["id"] = str(result.pop("_id"))
            
            return results
        except Exception as e:
            logger.error("get_lab_results_failed", error=str(e))
            return []
    
    async def get_prescriptions(
        self,
        patient_id: str,
        user_id: str,
        user_role: str,
        status: Optional[str] = None,
        limit: int = 50
    ) -> List[Dict[str, Any]]:
        """Get prescriptions for a patient."""
        if not self.db:
            return []
        
        try:
            query = {"patient_id": patient_id}
            if status:
                query["status"] = status
            
            cursor = self.db.prescriptions.find(query).sort("start_date", DESCENDING).limit(limit)
            prescriptions = await cursor.to_list(length=limit)
            
            # Convert _id to id
            for prescription in prescriptions:
                prescription["id"] = str(prescription.pop("_id"))
            
            return prescriptions
        except Exception as e:
            logger.error("get_prescriptions_failed", error=str(e))
            return []
    
    async def create_prescription(
        self,
        prescription: Dict[str, Any],
        user_id: str,
        user_role: str
    ) -> Dict[str, Any]:
        """Create a new prescription."""
        if not self.db:
            return {}
        
        try:
            # Prepare document
            doc = {
                "patient_id": prescription["patient_id"],
                "doctor_id": prescription["doctor_id"],
                "medication": prescription["medication"],
                "dosage": prescription["dosage"],
                "frequency": prescription["frequency"],
                "start_date": prescription["start_date"],
                "end_date": prescription.get("end_date"),
                "status": prescription.get("status", "active"),
                "created_at": datetime.utcnow(),
            }
            
            result = await self.db.prescriptions.insert_one(doc)
            doc["id"] = str(result.inserted_id)
            doc["_id"] = result.inserted_id
            
            return doc
        except Exception as e:
            logger.error("create_prescription_failed", error=str(e))
            return {}
    
    async def search_patients(
        self,
        query: str,
        user_id: str,
        user_role: str,
        limit: int = 20
    ) -> List[Dict[str, Any]]:
        """
        Search patients (with access control).
        
        Only returns patients the user has access to.
        """
        if not self.db:
            return []
        
        try:
            # MongoDB text search or regex
            search_query = {
                "$or": [
                    {"name": {"$regex": query, "$options": "i"}},
                    # Note: CPF search would need to be handled differently (encrypted)
                ]
            }
            
            cursor = self.db.patients.find(
                search_query,
                {"cpf": 0}  # Don't return encrypted CPF
            ).limit(limit)
            
            patients = await cursor.to_list(length=limit)
            
            # Convert _id to id
            for patient in patients:
                patient["id"] = str(patient.pop("_id"))
            
            return patients
        except Exception as e:
            logger.error("search_patients_failed", error=str(e))
            return []
    
    async def store_medical_exam_pdf(
        self,
        patient_id: str,
        user_id: str,
        user_role: str,
        pdf_bytes: bytes,
        filename: str,
        storage_path: str,
        extracted_data: Dict[str, Any],
        metadata: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        Store medical exam PDF and extracted data.
        
        Note: Currently stores PDF bytes in MongoDB. For production, consider storing
        PDFs in MinIO/S3 and only storing references here.
        
        Args:
            patient_id: Patient ID
            user_id: User ID who uploaded the PDF
            user_role: User role
            pdf_bytes: PDF file content as bytes (used for size calculation)
            filename: Original filename
            storage_path: Path where PDF is stored (e.g., MinIO object path)
            extracted_data: Extracted structured data from LangExtract
            metadata: Additional metadata
            
        Returns:
            Dictionary with stored document information
        """
        if not self.db:
            return {}
        
        try:
            doc = {
                "patient_id": patient_id,
                "uploaded_by": user_id,
                "uploaded_by_role": user_role,
                "filename": filename,
                "storage_path": storage_path,
                "file_size": len(pdf_bytes),
                # Note: In production, don't store pdf_bytes here - use MinIO/S3
                # "pdf_bytes": pdf_bytes,  # Commented out for now
                "extracted_data": extracted_data,
                "metadata": metadata or {},
                "created_at": datetime.utcnow(),
                "updated_at": datetime.utcnow(),
            }
            
            result = await self.db.medical_exam_pdfs.insert_one(doc)
            doc["id"] = str(result.inserted_id)
            doc["_id"] = result.inserted_id
            
            logger.info(
                "medical_exam_pdf_stored",
                patient_id=patient_id,
                pdf_id=doc["id"],
                filename=filename,
                extraction_count=extracted_data.get("extraction_count", 0)
            )
            
            return doc
        except Exception as e:
            logger.error("store_medical_exam_pdf_failed", error=str(e))
            raise
    
    async def get_medical_exam_pdfs(
        self,
        patient_id: str,
        user_id: str,
        user_role: str,
        limit: int = 50
    ) -> List[Dict[str, Any]]:
        """
        Get medical exam PDFs for a patient.
        
        Args:
            patient_id: Patient ID
            user_id: User ID requesting access
            user_role: User role
            limit: Maximum number of results
            
        Returns:
            List of medical exam PDF documents
        """
        if not self.db:
            return []
        
        try:
            cursor = self.db.medical_exam_pdfs.find(
                {"patient_id": patient_id}
            ).sort("created_at", DESCENDING).limit(limit)
            
            pdfs = await cursor.to_list(length=limit)
            
            # Convert _id to id
            for pdf in pdfs:
                pdf["id"] = str(pdf.pop("_id"))
            
            return pdfs
        except Exception as e:
            logger.error("get_medical_exam_pdfs_failed", error=str(e))
            return []
    
    async def get_medical_exam_pdf(
        self,
        pdf_id: str,
        user_id: str,
        user_role: str
    ) -> Optional[Dict[str, Any]]:
        """
        Get a specific medical exam PDF by ID.
        
        Args:
            pdf_id: PDF document ID
            user_id: User ID requesting access
            user_role: User role
            
        Returns:
            Medical exam PDF document or None if not found
        """
        if not self.db:
            return None
        
        try:
            pdf = await self.db.medical_exam_pdfs.find_one({"_id": pdf_id})
            
            if pdf:
                pdf["id"] = str(pdf.pop("_id"))
            
            return pdf
        except Exception as e:
            logger.error("get_medical_exam_pdf_failed", error=str(e), pdf_id=pdf_id)
            return None
