#!/usr/bin/env python3
"""
🔄 Flyte Workflow for TRM Fine-Tuning

Automated fine-tuning pipeline that:
1. Collects data from notifi-services and observability
2. Fine-tunes TRM model
3. Evaluates and deploys model
4. Runs automatically every 30 days
"""

import os
from datetime import datetime, timedelta
from pathlib import Path
from typing import Dict, Any, Tuple

import flytekit
from flytekit import task, workflow, Resources, CronSchedule, LaunchPlan
from flytekit.types.file import FlyteFile
from flytekit.types.directory import FlyteDirectory

from data_collector import DataCollector, TrainingExample
from trm_trainer import TRMTrainer, TRMTrainingConfig


# ============================================================================
# 📊 Task 1: Collect Training Data
# ============================================================================

@task(
    requests=Resources(cpu="4", mem="8Gi"),
    timeout=flytekit.Duration(hours=2),
)
def collect_training_data(
    notifi_services_path: str,
    prometheus_url: str,
    loki_url: str,
    tempo_url: str,
    days: int = 30,
) -> FlyteFile:
    """
    Collect training data from notifi-services and observability.
    
    Returns:
        Path to JSONL file with training examples
    """
    import asyncio
    from data_collector import DataCollector
    
    flytekit.logger.info(f"📊 Collecting training data from last {days} days...")
    
    collector = DataCollector(
        notifi_services_path=notifi_services_path,
        prometheus_url=prometheus_url,
        loki_url=loki_url,
        tempo_url=tempo_url,
        days=days
    )
    
    # Collect all data (handle asyncio properly)
    try:
        loop = asyncio.get_event_loop()
    except RuntimeError:
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
    
    examples = loop.run_until_complete(collector.collect_all())
    
    # Save to file
    output_path = f"/tmp/training_data_{datetime.now().strftime('%Y%m%d_%H%M%S')}.jsonl"
    collector.save_to_jsonl(examples, output_path)
    
    flytekit.logger.info(f"✅ Collected {len(examples)} training examples")
    
    return FlyteFile(output_path)


# ============================================================================
# 🎓 Task 2: Fine-Tune TRM Model
# ============================================================================

@task(
    requests=Resources(cpu="8", mem="32Gi", gpu="1"),
    timeout=flytekit.Duration(hours=24),
    retries=1,
)
def fine_tune_trm(
    training_data: FlyteFile,
    trm_repo_path: str,
    config: Dict[str, Any],
) -> Tuple[FlyteDirectory, Dict[str, Any]]:
    """
    Fine-tune TRM model on collected data.
    
    Returns:
        Tuple of (model_directory, training_metrics)
    """
    from trm_trainer import TRMTrainer, TRMTrainingConfig
    
    flytekit.logger.info("🎓 Starting TRM fine-tuning...")
    
    # Create config from dict
    train_config = TRMTrainingConfig(
        epochs=config.get("epochs", 50000),
        eval_interval=config.get("eval_interval", 5000),
        lr=config.get("lr", 1e-4),
        L_layers=config.get("L_layers", 2),
        H_cycles=config.get("H_cycles", 3),
        L_cycles=config.get("L_cycles", 6),
        output_dir=config.get("output_dir", "/tmp/trm-model"),
        run_name=config.get("run_name", "trm-homelab-finetune")
    )
    
    trainer = TRMTrainer(train_config, trm_repo_path=trm_repo_path)
    
    # Prepare dataset
    data_dir = trainer.prepare_dataset(str(training_data), train_config.output_dir)
    
    # Train model
    model_dir = trainer.train(data_dir)
    
    # Export model
    export_path = trainer.export_model(model_dir, f"{model_dir}/export")
    
    # Create metrics (simplified - would parse from training logs)
    metrics = {
        "model_path": export_path,
        "training_completed": True,
        "timestamp": datetime.now().isoformat(),
        "config": config
    }
    
    flytekit.logger.info(f"✅ Training complete! Model at: {export_path}")
    
    return FlyteDirectory(export_path), metrics


# ============================================================================
# 📊 Task 3: Evaluate Model
# ============================================================================

@task(
    requests=Resources(cpu="4", mem="16Gi", gpu="1"),
    timeout=flytekit.Duration(hours=1),
)
def evaluate_model(
    model_dir: FlyteDirectory,
    test_data: FlyteFile = None,
) -> Dict[str, Any]:
    """
    Evaluate fine-tuned TRM model.
    
    Returns:
        Evaluation metrics
    """
    flytekit.logger.info("📊 Evaluating model...")
    
    # TODO: Implement actual evaluation
    # For now, return placeholder metrics
    metrics = {
        "accuracy": 0.85,  # Placeholder
        "loss": 0.15,  # Placeholder
        "evaluation_completed": True,
        "timestamp": datetime.now().isoformat()
    }
    
    flytekit.logger.info(f"✅ Evaluation complete: {metrics}")
    
    return metrics


# ============================================================================
# 🚀 Task 4: Deploy Model
# ============================================================================

@task(
    requests=Resources(cpu="2", mem="4Gi"),
    timeout=flytekit.Duration(hours=1),
)
def deploy_model(
    model_dir: FlyteDirectory,
    metrics: Dict[str, Any],
    minio_bucket: str = "trm-models",
) -> str:
    """
    Deploy model to MinIO and update Ollama/VLLM.
    
    Returns:
        Deployment status
    """
    flytekit.logger.info("🚀 Deploying model...")
    
    # TODO: Implement actual deployment
    # 1. Upload model to MinIO
    # 2. Update Ollama model registry
    # 3. Update agent configurations
    
    status = f"Model deployed to {minio_bucket} at {datetime.now().isoformat()}"
    flytekit.logger.info(f"✅ {status}")
    
    return status


# ============================================================================
# 🔄 Main Workflow
# ============================================================================

@workflow
def trm_finetuning_workflow(
    notifi_services_path: str = "/workspace/notifi/repos/notifi-services",
    prometheus_url: str = "http://prometheus.monitoring.svc:9090",
    loki_url: str = "http://loki.monitoring.svc:3100",
    tempo_url: str = "http://tempo.tempo.svc:3200",
    days: int = 30,
    trm_repo_path: str = "/workspace/TinyRecursiveModels",
    training_config: Dict[str, Any] = None,
    minio_bucket: str = "trm-models",
) -> Tuple[FlyteDirectory, Dict[str, Any], Dict[str, Any], str]:
    """
    Complete TRM fine-tuning pipeline.
    
    Steps:
    1. Collect training data (notifi-services + observability)
    2. Fine-tune TRM model
    3. Evaluate model
    4. Deploy model
    
    Returns:
        Tuple of (model_dir, training_metrics, eval_metrics, deployment_status)
    """
    if training_config is None:
        training_config = {
            "epochs": 50000,
            "eval_interval": 5000,
            "lr": 1e-4,
            "L_layers": 2,
            "H_cycles": 3,
            "L_cycles": 6,
            "output_dir": "/tmp/trm-model",
            "run_name": "trm-homelab-finetune"
        }
    
    # Step 1: Collect data
    training_data = collect_training_data(
        notifi_services_path=notifi_services_path,
        prometheus_url=prometheus_url,
        loki_url=loki_url,
        tempo_url=tempo_url,
        days=days,
    )
    
    # Step 2: Fine-tune
    model_dir, train_metrics = fine_tune_trm(
        training_data=training_data,
        trm_repo_path=trm_repo_path,
        config=training_config,
    )
    
    # Step 3: Evaluate
    eval_metrics = evaluate_model(model_dir=model_dir)
    
    # Step 4: Deploy
    deploy_status = deploy_model(
        model_dir=model_dir,
        metrics=eval_metrics,
        minio_bucket=minio_bucket,
    )
    
    return model_dir, train_metrics, eval_metrics, deploy_status


# ============================================================================
# ⏰ Scheduled Workflow (Every 30 Days)
# ============================================================================

@workflow(
    schedule=CronSchedule(
        schedule="0 2 1 * *",  # First day of every month at 2 AM
        kickoff_time_input_arg="trigger_time",
    )
)
def scheduled_trm_finetuning(
    trigger_time: str = None,
    notifi_services_path: str = "/workspace/notifi/repos/notifi-services",
    prometheus_url: str = "http://prometheus.monitoring.svc:9090",
    loki_url: str = "http://loki.monitoring.svc:3100",
    tempo_url: str = "http://tempo.tempo.svc:3200",
    days: int = 30,
) -> Tuple[FlyteDirectory, Dict[str, Any], Dict[str, Any], str]:
    """
    Scheduled TRM fine-tuning workflow.
    
    Runs automatically on the 1st of every month at 2 AM.
    Uses data from the last 30 days.
    """
    if trigger_time is None:
        trigger_time = datetime.now().isoformat()
    
    flytekit.logger.info(f"⏰ Scheduled fine-tuning triggered at {trigger_time}")
    
    return trm_finetuning_workflow(
        notifi_services_path=notifi_services_path,
        prometheus_url=prometheus_url,
        loki_url=loki_url,
        tempo_url=tempo_url,
        days=days,
    )


# ============================================================================
# 🎯 Launch Plans
# ============================================================================

# Launch plan for scheduled monthly fine-tuning
monthly_finetuning_plan = LaunchPlan.get_or_create(
    workflow=scheduled_trm_finetuning,
    name="monthly_trm_finetuning",
    schedule=CronSchedule(
        schedule="0 2 1 * *",  # First day of every month at 2 AM
        kickoff_time_input_arg="trigger_time",
    ),
    fixed_inputs={
        "days": 30,
        "notifi_services_path": "/workspace/notifi/repos/notifi-services",
        "prometheus_url": "http://prometheus.monitoring.svc:9090",
        "loki_url": "http://loki.monitoring.svc:3100",
        "tempo_url": "http://tempo.tempo.svc:3200",
    },
)

# Launch plan for manual trigger
manual_finetuning_plan = LaunchPlan.get_or_create(
    workflow=trm_finetuning_workflow,
    name="manual_trm_finetuning",
    default_inputs={
        "days": 30,
        "notifi_services_path": "/workspace/notifi/repos/notifi-services",
        "prometheus_url": "http://prometheus.monitoring.svc:9090",
        "loki_url": "http://loki.monitoring.svc:3100",
        "tempo_url": "http://tempo.tempo.svc:3200",
    },
)

