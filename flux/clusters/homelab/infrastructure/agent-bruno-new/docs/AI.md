# AI/ML Foundations Guide for Agent Bruno

> **Purpose**: This comprehensive guide provides the foundational AI/ML concepts you need to understand before diving into the Agent Bruno implementation. Each section connects theory to practical implementation with real-world examples.

**📖 Reading Guide**:
- **Beginners**: Read sections 1-5 sequentially to build from basics to RAG systems
- **ML Engineers**: Jump to sections 6-8 for production ML topics and MLOps
- **SRE/DevOps**: Focus on sections 8-10 for observability, deployment, and infrastructure
- **Quick Reference**: Each section links to relevant Agent Bruno documentation and includes practical code examples

**🎯 Learning Objectives**:
- Understand neural network fundamentals and training processes
- Master embeddings, vector representations, and similarity search
- Learn about LLMs, fine-tuning, and parameter-efficient methods
- Implement production-ready RAG systems with hybrid retrieval
- Design ML pipelines with proper observability and monitoring
- Build event-driven AI architectures for scalable systems

---

## Table of Contents

1. [Neural Networks Fundamentals](#1-neural-networks-fundamentals)
2. [Loss Functions & Training](#2-loss-functions--training)
3. [Embeddings & Vector Representations](#3-embeddings--vector-representations)
4. [Language Models (LLMs)](#4-language-models-llms)
5. [Retrieval-Augmented Generation (RAG)](#5-retrieval-augmented-generation-rag)
6. [Fine-tuning & LoRA](#6-fine-tuning--lora)
7. [Vector Databases](#7-vector-databases)
8. [Production ML Engineering](#8-production-ml-engineering)
9. [AI Observability](#9-ai-observability)
10. [Event-Driven AI Architectures](#10-event-driven-ai-architectures)

---

## 1. Neural Networks Fundamentals

### What is a Neural Network?

A neural network is a computational model inspired by biological neurons that learns to map inputs to outputs through training. Modern neural networks can have millions or billions of parameters and are the foundation of all AI systems.

**Key Concepts**:

```
┌─────────────────────────────────────────────────┐
│           Basic Neural Network                  │
├─────────────────────────────────────────────────┤
│                                                 │
│  Input Layer  →  Hidden Layers  →  Output Layer│
│      ↓              ↓                   ↓       │
│   Features      Patterns          Predictions   │
│                                                 │
│  Example: Text Embedding                        │
│  Input: "Hello world"                           │
│  Hidden: [0.2, -0.5, 0.8, ...]  (vector)        │
│  Output: Semantic representation                │
└─────────────────────────────────────────────────┘
```

### Neural Network Architectures

#### 1. Feedforward Networks (MLPs)
Basic architecture where information flows in one direction:

```python
import torch
import torch.nn as nn

class MLP(nn.Module):
    def __init__(self, input_size, hidden_size, output_size):
        super().__init__()
        self.layers = nn.Sequential(
            nn.Linear(input_size, hidden_size),
            nn.ReLU(),
            nn.Linear(hidden_size, hidden_size),
            nn.ReLU(),
            nn.Linear(hidden_size, output_size)
        )
    
    def forward(self, x):
        return self.layers(x)
```

#### 2. Convolutional Neural Networks (CNNs)
Specialized for spatial data like images:

```python
class CNN(nn.Module):
    def __init__(self, num_classes):
        super().__init__()
        self.conv_layers = nn.Sequential(
            nn.Conv2d(3, 64, kernel_size=3, padding=1),
            nn.ReLU(),
            nn.MaxPool2d(2),
            nn.Conv2d(64, 128, kernel_size=3, padding=1),
            nn.ReLU(),
            nn.MaxPool2d(2),
        )
        self.classifier = nn.Sequential(
            nn.Flatten(),
            nn.Linear(128 * 8 * 8, 512),
            nn.ReLU(),
            nn.Linear(512, num_classes)
        )
    
    def forward(self, x):
        x = self.conv_layers(x)
        return self.classifier(x)
```

#### 3. Recurrent Neural Networks (RNNs)
Designed for sequential data:

```python
class RNN(nn.Module):
    def __init__(self, input_size, hidden_size, output_size):
        super().__init__()
        self.hidden_size = hidden_size
        self.rnn = nn.RNN(input_size, hidden_size, batch_first=True)
        self.fc = nn.Linear(hidden_size, output_size)
    
    def forward(self, x):
        output, hidden = self.rnn(x)
        return self.fc(output[:, -1, :])  # Last timestep
```

#### 4. Transformers (The Foundation of Modern LLMs)
Revolutionary architecture using attention mechanisms:

```python
class TransformerBlock(nn.Module):
    def __init__(self, d_model, n_heads, d_ff):
        super().__init__()
        self.attention = nn.MultiheadAttention(d_model, n_heads)
        self.norm1 = nn.LayerNorm(d_model)
        self.norm2 = nn.LayerNorm(d_model)
        self.ff = nn.Sequential(
            nn.Linear(d_model, d_ff),
            nn.ReLU(),
            nn.Linear(d_ff, d_model)
        )
    
    def forward(self, x):
        # Self-attention
        attn_output, _ = self.attention(x, x, x)
        x = self.norm1(x + attn_output)
        
        # Feed-forward
        ff_output = self.ff(x)
        x = self.norm2(x + ff_output)
        return x
```

### Key Components

1. **Layers**: Stacks of neurons that transform input progressively
2. **Weights**: Learnable parameters that determine transformations
3. **Activations**: Non-linear functions (ReLU, sigmoid, tanh, GELU)
4. **Forward Pass**: Data flows through network to produce output
5. **Backward Pass**: Gradients flow back to update weights
6. **Attention Mechanisms**: Allow models to focus on relevant parts of input
7. **Residual Connections**: Help with gradient flow in deep networks

### Activation Functions

```python
# Common activation functions
import torch.nn.functional as F

# ReLU: f(x) = max(0, x) - Most common, simple, fast
x_relu = F.relu(x)

# GELU: f(x) = x * Φ(x) - Used in Transformers, smoother than ReLU
x_gelu = F.gelu(x)

# Sigmoid: f(x) = 1/(1 + e^(-x)) - Outputs between 0 and 1
x_sigmoid = torch.sigmoid(x)

# Tanh: f(x) = (e^x - e^(-x))/(e^x + e^(-x)) - Outputs between -1 and 1
x_tanh = torch.tanh(x)
```

**In Agent Bruno**: 
- **Embedding Model**: Uses transformer-based architecture (e.g., sentence-transformers) to convert text to vectors
- **LLM (Ollama)**: Uses transformer architecture for text generation
- **Hybrid RAG**: Combines semantic (transformer-based) and keyword search
- **Memory System**: Uses vector embeddings for long-term memory storage

📖 **Related Docs**: [ARCHITECTURE.md](./ARCHITECTURE.md#technology-stack) | [RAG.md](./RAG.md)

---

## 2. Loss Functions & Training

### What is a Loss Function?

A **loss function** (also called a cost function or objective function) measures how far your neural network's predictions are from the actual target values. It quantifies the "error" or "cost" of your model's predictions.

### The Training Loop

Loss functions are central to the training process:

```
┌─────────────────────────────────────────────────┐
│         Neural Network Training Loop            │
├─────────────────────────────────────────────────┤
│  1. Forward Pass                                │
│     Input → Network → Prediction                │
│                                                 │
│  2. Loss Calculation  ← Measures Error          │
│     Loss = loss_function(prediction, target)    │
│                                                 │
│  3. Backward Pass (Backpropagation)             │
│     Compute gradients: ∂Loss/∂weights           │
│                                                 │
│  4. Weight Update (Optimization)                │
│     weights = weights - learning_rate × gradient│
│                                                 │
│     Repeat...                                   │
└─────────────────────────────────────────────────┘
```

## The Training Process in Detail

### 1. Forward Pass

The forward pass is where your input data flows through the neural network layer by layer to produce a prediction.

**How it works:**
- Start with input data (e.g., an image, text, numbers)
- Pass it through the first layer: `output₁ = activation(weights₁ × input + bias₁)`
- Each subsequent layer takes the previous layer's output as input
- Continue until you reach the final output layer
- The final output is your prediction (ŷ)

**Example:**
```python
# Input: image of a cat
x = input_image  # shape: [batch_size, 3, 224, 224]

# Layer 1: Convolutional layer
x = conv1(x)      # shape: [batch_size, 64, 112, 112]
x = relu(x)       # activation function

# Layer 2: Another conv layer
x = conv2(x)      # shape: [batch_size, 128, 56, 56]
x = relu(x)

# ... more layers ...

# Final layer: Classification
logits = fc(x)    # shape: [batch_size, num_classes]
predictions = softmax(logits)  # probabilities for each class
```

**Key Point:** During forward pass, we're just computing outputs - no learning happens yet!

### 2. Loss Calculation

After getting predictions, we need to measure how wrong they are. This is what the loss function does.

**How it works:**
- Compare your prediction (ŷ) with the actual target (y)
- Compute a single number that represents the error
- Lower loss = better predictions
- Higher loss = worse predictions

**Example:**
```python
# For classification (cat vs dog)
predictions = model(image)     # [0.9, 0.1] (90% cat, 10% dog)
target = [1, 0]                # actual label: cat

# Cross-entropy loss
loss = -sum(target * log(predictions))
# loss = -(1 × log(0.9) + 0 × log(0.1))
# loss ≈ 0.105 (pretty good!)

# If prediction was wrong [0.2, 0.8] (20% cat, 80% dog)
# loss = -(1 × log(0.2) + 0 × log(0.8))
# loss ≈ 1.609 (very bad!)
```

**Key Point:** The loss is a single number that tells us "how bad" our current model is. Our goal is to minimize this number.

### 3. Backward Pass (Backpropagation)

Now comes the magic! We need to figure out how to adjust each weight to reduce the loss.

**How it works:**
- Calculate the gradient (derivative) of the loss with respect to each weight
- Gradient tells us: "If I increase this weight, how much will the loss change?"
- Use the chain rule to propagate gradients backward through the network
- Start from the output and work backward to the input

**Mathematical concept:**
```
∂Loss/∂weight = how much loss changes when we change this weight

If gradient is positive: increasing weight increases loss (bad!)
If gradient is negative: increasing weight decreases loss (good!)
If gradient is large: weight has big impact on loss
If gradient is small: weight has small impact on loss
```

**Example:**
```python
# PyTorch automatically computes gradients
loss.backward()  # Compute all gradients!

# Now each parameter has a gradient
for name, param in model.named_parameters():
    print(f"{name}: gradient = {param.grad}")
    # conv1.weight: gradient = tensor([[-0.001, 0.003, ...]])
    # conv1.bias: gradient = tensor([0.002, -0.001, ...])
```

**Visual example:**
```
Output Layer (Loss = 0.5)
    ↓ ← compute gradient here first
Hidden Layer 2
    ↓ ← then propagate backward
Hidden Layer 1
    ↓ ← keep going backward
Input Layer
```

**Key Point:** Backpropagation efficiently computes how each weight contributed to the final loss. It's the core algorithm that makes deep learning possible!

### 4. Weight Update (Optimization)

Finally, we use the gradients to update our weights to make the model better.

**How it works:**
- Take each weight and nudge it in the opposite direction of its gradient
- The gradient tells us which way increases loss, so we go the opposite way
- Learning rate controls how big each step is

**The update rule:**
```
new_weight = old_weight - learning_rate × gradient

Example:
old_weight = 0.5
gradient = 0.2 (positive means this weight increased the loss)
learning_rate = 0.01

new_weight = 0.5 - (0.01 × 0.2)
new_weight = 0.5 - 0.002
new_weight = 0.498  ← slightly smaller, should reduce loss!
```

**Example in code:**
```python
# Simple gradient descent
learning_rate = 0.001

for param in model.parameters():
    param.data = param.data - learning_rate * param.grad

# Or use an optimizer (recommended)
optimizer = torch.optim.Adam(model.parameters(), lr=0.001)
optimizer.step()  # Updates all weights automatically
```

**Different optimizers:**
- **SGD (Stochastic Gradient Descent)**: Basic, just uses gradient
- **Adam**: Adaptive learning rate, momentum, very popular
- **RMSprop**: Adapts learning rate per parameter
- **AdaGrad**: Good for sparse data

**Key Point:** This is where actual learning happens! Weights change slightly to reduce the loss. After millions of these tiny updates, the model becomes accurate.

### 5. Repeat Until Convergence

The entire process (forward → loss → backward → update) repeats many times:

```python
for epoch in range(num_epochs):           # Train for multiple epochs
    for batch in dataloader:              # Process data in batches
        # 1. Forward pass
        predictions = model(batch.input)
        
        # 2. Loss calculation
        loss = criterion(predictions, batch.target)
        
        # 3. Backward pass
        optimizer.zero_grad()             # Clear old gradients
        loss.backward()                   # Compute new gradients
        
        # 4. Weight update
        optimizer.step()                  # Update weights
        
    print(f"Epoch {epoch}: Loss = {loss.item()}")
```

**When to stop:**
- Loss stops decreasing (converged)
- Validation loss starts increasing (overfitting)
- Reached maximum number of epochs
- Model performance is good enough

## Common Loss Functions

### For Regression (predicting continuous values)

**Mean Squared Error (MSE)**:
- Formula: L = (1/n) × Σ(yᵢ - ŷᵢ)²
- Heavily penalizes large errors (due to squaring)
- Most common for regression tasks

**Mean Absolute Error (MAE)**:
- Formula: L = (1/n) × Σ|yᵢ - ŷᵢ|
- More robust to outliers than MSE
- Treats all errors linearly

### For Classification

**Cross-Entropy Loss**:
- Measures difference between predicted probability distributions and true labels
- For multi-class: L = -Σ yᵢ × log(ŷᵢ)
- Encourages the model to output high probabilities for correct classes

**Binary Cross-Entropy**:
- For binary classification (yes/no, 0/1)
- Formula: L = -[y × log(ŷ) + (1-y) × log(1-ŷ)]

### For Language Models

**Cross-Entropy Loss (for next token prediction)**:
- Used in models like GPT, BERT, etc.
- Predicts the probability distribution over the entire vocabulary
- Minimizes the negative log-likelihood of the correct next token

## Practical Example

In a typical PyTorch training loop:

```python
# Forward pass
logits = model(input_ids)

# Loss calculation
loss = F.cross_entropy(logits.view(-1, vocab_size), targets.view(-1))

# Backward pass
loss.backward()

# Optimizer step
optimizer.step()
```

## Why Loss Functions Matter

The loss function is what makes learning possible:

- **Provides Direction**: Tells the optimizer which way to adjust weights
- **Measures Progress**: Lower loss = better predictions
- **Guides Training**: Different loss functions suit different tasks
- **Enables Backpropagation**: Gradients are computed from the loss

Without a loss function, the network wouldn't know if it's getting better or worse, and learning would be impossible!

---

## Complete Training Loop - Agent Bruno Context

Now let's see how all these pieces fit together in a complete training loop, using examples from Agent Bruno's stack (PyTorch, Ollama, LanceDB).

### Full Training Loop Implementation

```python
import torch
import torch.nn as nn
import torch.optim as optim
from torch.utils.data import DataLoader
import lancedb
from typing import Dict, List
import time

class AgentBrunoTrainer:
    """
    Complete training loop for Agent Bruno's embedding model.
    
    Demonstrates all training components:
    - Forward pass
    - Loss calculation  
    - Backward pass (backpropagation)
    - Optimizer step (weight update)
    - Validation
    - Checkpointing
    """
    
    def __init__(
        self,
        model: nn.Module,
        train_loader: DataLoader,
        val_loader: DataLoader,
        device: str = "cuda" if torch.cuda.is_available() else "cpu"
    ):
        self.model = model.to(device)
        self.train_loader = train_loader
        self.val_loader = val_loader
        self.device = device
        
        # Loss function for embedding model (contrastive learning)
        self.criterion = nn.TripletMarginLoss(margin=0.5)
        
        # Optimizer - AdamW is standard for transformers
        self.optimizer = optim.AdamW(
            model.parameters(),
            lr=2e-5,           # Learning rate
            betas=(0.9, 0.999), # Adam beta parameters
            weight_decay=0.01   # L2 regularization
        )
        
        # Learning rate scheduler (warmup + decay)
        self.scheduler = optim.lr_scheduler.OneCycleLR(
            self.optimizer,
            max_lr=2e-5,
            steps_per_epoch=len(train_loader),
            epochs=10
        )
        
        # Metrics tracking
        self.metrics = {
            'train_loss': [],
            'val_loss': [],
            'learning_rate': []
        }
    
    def train_epoch(self, epoch: int) -> float:
        """
        Train for one epoch.
        
        Returns:
            Average training loss for the epoch
        """
        self.model.train()  # Set to training mode (enables dropout, etc.)
        total_loss = 0.0
        num_batches = 0
        
        for batch_idx, batch in enumerate(self.train_loader):
            # ─────────────────────────────────────────────────
            # Step 1: FORWARD PASS
            # ─────────────────────────────────────────────────
            # Move data to device (GPU/CPU)
            anchor = batch['anchor'].to(self.device)
            positive = batch['positive'].to(self.device)
            negative = batch['negative'].to(self.device)
            
            # Forward pass through model
            anchor_emb = self.model(anchor['input_ids'], anchor['attention_mask'])
            positive_emb = self.model(positive['input_ids'], positive['attention_mask'])
            negative_emb = self.model(negative['input_ids'], negative['attention_mask'])
            
            # ─────────────────────────────────────────────────
            # Step 2: LOSS CALCULATION
            # ─────────────────────────────────────────────────
            # Triplet loss: anchor should be closer to positive than negative
            # Loss = max(0, margin + dist(anchor, positive) - dist(anchor, negative))
            loss = self.criterion(anchor_emb, positive_emb, negative_emb)
            
            # ─────────────────────────────────────────────────
            # Step 3: BACKWARD PASS (Backpropagation)
            # ─────────────────────────────────────────────────
            # Clear previous gradients
            self.optimizer.zero_grad()
            
            # Compute gradients (∂Loss/∂weights)
            loss.backward()
            
            # Gradient clipping (prevent exploding gradients)
            torch.nn.utils.clip_grad_norm_(self.model.parameters(), max_norm=1.0)
            
            # ─────────────────────────────────────────────────
            # Step 4: WEIGHT UPDATE (Optimization)
            # ─────────────────────────────────────────────────
            # Update weights: w = w - lr * gradient
            self.optimizer.step()
            
            # Update learning rate (scheduler step)
            self.scheduler.step()
            
            # ─────────────────────────────────────────────────
            # Metrics tracking
            # ─────────────────────────────────────────────────
            total_loss += loss.item()
            num_batches += 1
            
            # Log every 100 batches
            if batch_idx % 100 == 0:
                current_lr = self.scheduler.get_last_lr()[0]
                print(f"Epoch [{epoch}] Batch [{batch_idx}/{len(self.train_loader)}] "
                      f"Loss: {loss.item():.4f} LR: {current_lr:.2e}")
        
        avg_loss = total_loss / num_batches
        return avg_loss
    
    def validate(self) -> float:
        """
        Validation loop (no gradient computation).
        
        Returns:
            Average validation loss
        """
        self.model.eval()  # Set to evaluation mode (disables dropout)
        total_loss = 0.0
        num_batches = 0
        
        # Disable gradient computation for validation
        with torch.no_grad():
            for batch in self.val_loader:
                anchor = batch['anchor'].to(self.device)
                positive = batch['positive'].to(self.device)
                negative = batch['negative'].to(self.device)
                
                # Forward pass only (no backward pass)
                anchor_emb = self.model(anchor['input_ids'], anchor['attention_mask'])
                positive_emb = self.model(positive['input_ids'], positive['attention_mask'])
                negative_emb = self.model(negative['input_ids'], negative['attention_mask'])
                
                # Calculate loss
                loss = self.criterion(anchor_emb, positive_emb, negative_emb)
                
                total_loss += loss.item()
                num_batches += 1
        
        avg_loss = total_loss / num_batches
        return avg_loss
    
    def train(self, num_epochs: int, checkpoint_dir: str = "./checkpoints"):
        """
        Full training loop for multiple epochs.
        """
        print(f"Training on {self.device}")
        print(f"Model parameters: {sum(p.numel() for p in self.model.parameters()):,}")
        
        best_val_loss = float('inf')
        
        for epoch in range(num_epochs):
            epoch_start = time.time()
            
            # Train for one epoch
            train_loss = self.train_epoch(epoch)
            
            # Validate
            val_loss = self.validate()
            
            # Track metrics
            self.metrics['train_loss'].append(train_loss)
            self.metrics['val_loss'].append(val_loss)
            self.metrics['learning_rate'].append(self.scheduler.get_last_lr()[0])
            
            epoch_time = time.time() - epoch_start
            
            print(f"\nEpoch [{epoch+1}/{num_epochs}] Complete:")
            print(f"  Train Loss: {train_loss:.4f}")
            print(f"  Val Loss:   {val_loss:.4f}")
            print(f"  Time:       {epoch_time:.2f}s")
            print(f"  LR:         {self.scheduler.get_last_lr()[0]:.2e}\n")
            
            # Save best model
            if val_loss < best_val_loss:
                best_val_loss = val_loss
                self.save_checkpoint(
                    f"{checkpoint_dir}/best_model.pt",
                    epoch,
                    val_loss,
                    is_best=True
                )
                print(f"  ✓ New best model saved (val_loss: {val_loss:.4f})")
            
            # Save regular checkpoint
            if (epoch + 1) % 5 == 0:  # Every 5 epochs
                self.save_checkpoint(
                    f"{checkpoint_dir}/checkpoint_epoch_{epoch+1}.pt",
                    epoch,
                    val_loss
                )
        
        print(f"\nTraining complete! Best val loss: {best_val_loss:.4f}")
        return self.metrics
    
    def save_checkpoint(self, path: str, epoch: int, val_loss: float, is_best: bool = False):
        """Save model checkpoint"""
        checkpoint = {
            'epoch': epoch,
            'model_state_dict': self.model.state_dict(),
            'optimizer_state_dict': self.optimizer.state_dict(),
            'scheduler_state_dict': self.scheduler.state_dict(),
            'val_loss': val_loss,
            'metrics': self.metrics
        }
        torch.save(checkpoint, path)
        
        if is_best:
            print(f"  💾 Saved best checkpoint to {path}")

# Usage example for Agent Bruno
if __name__ == "__main__":
    from embedding_model import EmbeddingModel
    from data_loader import get_data_loaders
    
    # Initialize model
    model = EmbeddingModel(vocab_size=30522, hidden_size=768)
    
    # Get data loaders
    train_loader, val_loader = get_data_loaders(batch_size=32)
    
    # Create trainer
    trainer = AgentBrunoTrainer(
        model=model,
        train_loader=train_loader,
        val_loader=val_loader
    )
    
    # Train model
    metrics = trainer.train(num_epochs=10)
    
    # After training, save embeddings to LanceDB
    print("\nSaving embeddings to LanceDB...")
    # ... (see Vector Database section)
```

### Training Loop Breakdown

Let's examine each component in detail:

#### 1. Forward Pass - Detailed View

```python
def forward_pass_detailed(self, input_data):
    """
    Detailed view of what happens during forward pass.
    """
    # Input data
    print(f"Input shape: {input_data.shape}")
    # Example: torch.Size([32, 768]) - 32 samples, 768 features
    
    # Layer 1: Linear transformation
    x = self.layer1(input_data)
    print(f"After layer1: {x.shape}")
    # torch.Size([32, 256]) - 32 samples, 256 features
    # Computation: x = W₁ × input + b₁
    # W₁.shape: (256, 768), b₁.shape: (256,)
    
    # Activation: ReLU(x) = max(0, x)
    x = torch.relu(x)
    print(f"After ReLU: same shape, but negatives → 0")
    
    # Layer 2: Another linear transformation
    output = self.layer2(x)
    print(f"Output shape: {output.shape}")
    # torch.Size([32, 10]) - 32 samples, 10 classes
    
    return output

# What's happening internally:
# 1. Matrix multiplication: y = xW^T + b
# 2. Activation function applied element-wise
# 3. Data shape transforms through the network
# 4. Final output = predictions
```

#### 2. Loss Calculation - Different Loss Functions

```python
# Classification: Cross-Entropy Loss
criterion_cls = nn.CrossEntropyLoss()
logits = model(input_ids)  # Raw scores
targets = torch.tensor([0, 1, 2, ...])  # Class labels
loss_cls = criterion_cls(logits, targets)
# Formula: -sum(y_true * log(softmax(logits)))

# Regression: Mean Squared Error (MSE)
criterion_reg = nn.MSELoss()
predictions = model(features)
targets = torch.tensor([1.5, 2.3, 0.8, ...])
loss_reg = criterion_reg(predictions, targets)
# Formula: mean((predictions - targets)²)

# Embeddings: Triplet Margin Loss (Agent Bruno)
criterion_emb = nn.TripletMarginLoss(margin=0.5)
anchor_emb = model(anchor)
positive_emb = model(positive)  # Similar to anchor
negative_emb = model(negative)  # Different from anchor
loss_emb = criterion_emb(anchor_emb, positive_emb, negative_emb)
# Formula: max(0, margin + dist(anchor, positive) - dist(anchor, negative))
# Goal: Make anchor closer to positive than to negative

# Contrastive Loss (Alternative for embeddings)
class ContrastiveLoss(nn.Module):
    def __init__(self, temperature=0.07):
        super().__init__()
        self.temperature = temperature
    
    def forward(self, embeddings, labels):
        """
        InfoNCE loss for contrastive learning.
        Used in SimCLR, CLIP, etc.
        """
        # Normalize embeddings
        embeddings = F.normalize(embeddings, p=2, dim=1)
        
        # Compute similarity matrix
        similarity = torch.matmul(embeddings, embeddings.T) / self.temperature
        
        # Mask out self-similarity
        mask = torch.eye(len(embeddings), device=embeddings.device).bool()
        similarity.masked_fill_(mask, -float('inf'))
        
        # Cross-entropy loss
        loss = F.cross_entropy(similarity, labels)
        
        return loss
```

#### 3. Backward Pass - Gradient Flow

```python
def visualize_gradients(model, loss):
    """
    Visualize how gradients flow backward through the network.
    """
    print("\n=== Gradient Flow ===")
    
    # Compute gradients
    loss.backward()
    
    # Check gradient magnitudes for each layer
    for name, param in model.named_parameters():
        if param.grad is not None:
            grad_norm = param.grad.norm().item()
            param_norm = param.data.norm().item()
            print(f"{name:30s} | Grad: {grad_norm:8.4f} | Param: {param_norm:8.4f} | Ratio: {grad_norm/param_norm:.4f}")
    
    print("\nGradient Statistics:")
    print("  - Large gradients → layer has big impact on loss")
    print("  - Small gradients → layer has small impact (or vanishing gradient)")
    print("  - Ratio shows relative change magnitude\n")

# Example output:
# embedding.weight               | Grad:   2.3456 | Param:  45.2134 | Ratio: 0.0519
# encoder.layer.0.weight         | Grad:   1.2345 | Param:  23.4567 | Ratio: 0.0526
# encoder.layer.1.weight         | Grad:   0.9876 | Param:  19.8765 | Ratio: 0.0497
# pooler.weight                  | Grad:   0.5432 | Param:  12.3456 | Ratio: 0.0440

# Gradient clipping (prevent exploding gradients)
torch.nn.utils.clip_grad_norm_(model.parameters(), max_norm=1.0)
print("Gradients clipped to max norm of 1.0")
```

#### 4. Optimizer Step - Different Optimizers

```python
# SGD (Stochastic Gradient Descent)
# Simplest: just move in direction of negative gradient
optimizer_sgd = optim.SGD(model.parameters(), lr=0.01, momentum=0.9)
# Update rule: w = w - lr * gradient + momentum * previous_update

# Adam (Adaptive Moment Estimation) 
# Most popular: adaptive learning rates per parameter
optimizer_adam = optim.Adam(model.parameters(), lr=0.001, betas=(0.9, 0.999))
# Uses running averages of gradients and squared gradients
# Automatically adjusts learning rate for each parameter

# AdamW (Adam with Weight Decay)
# Standard for transformers (BERT, GPT, etc.)
optimizer_adamw = optim.AdamW(
    model.parameters(),
    lr=2e-5,           # Small learning rate for fine-tuning
    betas=(0.9, 0.999),
    weight_decay=0.01   # L2 regularization
)
# Fixes weight decay implementation in Adam
# Used in Agent Bruno for embedding model training

# Optimizer step visualization
def visualize_optimizer_step(optimizer, model):
    """Show what happens during optimizer.step()"""
    
    # Before step
    print("Before optimizer step:")
    param_before = model.layer1.weight[0, 0].item()
    grad = model.layer1.weight.grad[0, 0].item()
    
    # Optimizer step
    optimizer.step()
    
    # After step
    param_after = model.layer1.weight[0, 0].item()
    change = param_after - param_before
    
    print(f"  Parameter: {param_before:.6f}")
    print(f"  Gradient:  {grad:.6f}")
    print(f"  New value: {param_after:.6f}")
    print(f"  Change:    {change:.6f}")
    print(f"  Expected:  {-0.001 * grad:.6f} (lr=0.001)")
```

### Learning Rate Scheduling

Learning rate scheduling is crucial for optimal training:

```python
from torch.optim.lr_scheduler import OneCycleLR, CosineAnnealingLR, StepLR

# 1. OneCycle Learning Rate (Agent Bruno uses this)
scheduler_onecycle = OneCycleLR(
    optimizer,
    max_lr=2e-5,        # Peak learning rate
    steps_per_epoch=len(train_loader),
    epochs=num_epochs,
    pct_start=0.3,      # Spend 30% of time warming up
    anneal_strategy='cos'
)

# 2. Cosine Annealing (smooth decay)
scheduler_cosine = CosineAnnealingLR(
    optimizer,
    T_max=num_epochs,   # Period of cosine cycle
    eta_min=1e-6        # Minimum learning rate
)

# 3. Step Decay (reduce at fixed intervals)
scheduler_step = StepLR(
    optimizer,
    step_size=5,        # Reduce every 5 epochs
    gamma=0.1           # Multiply by 0.1
)

# Learning rate schedule visualization
import matplotlib.pyplot as plt

lrs = []
for epoch in range(num_epochs):
    lrs.append(scheduler.get_last_lr()[0])
    scheduler.step()

plt.plot(lrs)
plt.xlabel('Epoch')
plt.ylabel('Learning Rate')
plt.title('Learning Rate Schedule')
plt.yscale('log')
plt.show()

# Typical schedule for Agent Bruno:
# Epochs 0-3:   Warmup  (lr increases from 0 → 2e-5)
# Epochs 3-7:   Peak    (lr stays at 2e-5)
# Epochs 7-10:  Decay   (lr decreases to 1e-6)
```

### Mixed Precision Training (Speed + Memory)

```python
from torch.cuda.amp import autocast, GradScaler

class MixedPrecisionTrainer:
    """
    Training with automatic mixed precision (FP16 + FP32).
    
    Benefits:
    - 2-3x faster training
    - 50% less memory usage
    - Minimal accuracy loss
    """
    
    def __init__(self, model, optimizer):
        self.model = model
        self.optimizer = optimizer
        self.scaler = GradScaler()  # Gradient scaling for FP16
    
    def train_step(self, batch):
        """Training step with mixed precision"""
        
        # Clear gradients
        self.optimizer.zero_grad()
        
        # Forward pass in FP16 (faster, less memory)
        with autocast():
            outputs = self.model(batch['input_ids'], batch['attention_mask'])
            loss = self.criterion(outputs, batch['labels'])
        
        # Backward pass with gradient scaling
        # (prevents underflow in FP16)
        self.scaler.scale(loss).backward()
        
        # Optimizer step with unscaling
        self.scaler.step(self.optimizer)
        self.scaler.update()
        
        return loss.item()

# Usage in Agent Bruno for faster training on GPU
if torch.cuda.is_available():
    trainer = MixedPrecisionTrainer(model, optimizer)
    print("Using mixed precision training (FP16)")
else:
    print("Using full precision training (FP32)")
```

---

## 3. Embeddings & Vector Representations

### What are Embeddings?

Embeddings are dense vector representations that capture the semantic meaning of data (text, images, etc.) in a high-dimensional space. They enable similarity search, clustering, and semantic understanding.

**Key Properties**:
- **Dense**: Every dimension has meaning (unlike sparse one-hot encodings)
- **Semantic**: Similar concepts have similar vectors
- **Dimensional**: Typically 128-1536 dimensions for text embeddings
- **Normalized**: Often unit vectors for efficient similarity computation

### Types of Embeddings

#### 1. Text Embeddings
Convert text to vectors that capture semantic meaning:

```python
from sentence_transformers import SentenceTransformer
import numpy as np

# Load a pre-trained embedding model
model = SentenceTransformer('all-MiniLM-L6-v2')

# Generate embeddings
texts = [
    "How to deploy a Kubernetes cluster",
    "Setting up a K8s deployment",
    "What's the weather like today?"
]

embeddings = model.encode(texts)
print(f"Embedding shape: {embeddings.shape}")  # (3, 384)

# Compute similarity
similarity = np.dot(embeddings[0], embeddings[1])  # High similarity
print(f"Similarity between K8s texts: {similarity:.3f}")
```

#### 2. Image Embeddings
Convert images to vectors using vision models:

```python
import torch
import torchvision.transforms as transforms
from torchvision.models import resnet50

# Load pre-trained ResNet
model = resnet50(pretrained=True)
model.eval()

# Preprocess image
transform = transforms.Compose([
    transforms.Resize(256),
    transforms.CenterCrop(224),
    transforms.ToTensor(),
    transforms.Normalize(mean=[0.485, 0.456, 0.406], 
                        std=[0.229, 0.224, 0.225])
])

def get_image_embedding(image_path):
    image = transform(Image.open(image_path)).unsqueeze(0)
    with torch.no_grad():
        features = model(image)
    return features.numpy()
```

#### 3. Multimodal Embeddings
Combine text and images in a shared embedding space:

```python
from transformers import CLIPProcessor, CLIPModel

model = CLIPModel.from_pretrained("openai/clip-vit-base-patch32")
processor = CLIPProcessor.from_pretrained("openai/clip-vit-base-patch32")

def get_multimodal_embedding(text, image):
    inputs = processor(text=[text], images=[image], return_tensors="pt", padding=True)
    outputs = model(**inputs)
    return outputs.text_embeds, outputs.image_embeds
```

### Embedding Models

#### Sentence Transformers
Popular for text embeddings with various models:

```python
from sentence_transformers import SentenceTransformer, util

# Different models for different use cases
models = {
    'general': 'all-MiniLM-L6-v2',           # Fast, general purpose
    'multilingual': 'paraphrase-multilingual-MiniLM-L12-v2',
    'code': 'microsoft/codebert-base',        # For code embeddings
    'medical': 'sentence-transformers/all-mpnet-base-v2'
}

# Example usage
model = SentenceTransformer('all-MiniLM-L6-v2')
embeddings = model.encode([
    "Kubernetes deployment failed",
    "Pod is not starting",
    "Service discovery issue"
])

# Find most similar
similarities = util.cos_sim(embeddings[0], embeddings[1:])
print(f"Similarity scores: {similarities}")
```

#### OpenAI Embeddings
High-quality embeddings from OpenAI:

```python
import openai

def get_openai_embedding(text, model="text-embedding-ada-002"):
    response = openai.Embedding.create(
        input=text,
        model=model
    )
    return response['data'][0]['embedding']

# Usage
embedding = get_openai_embedding("How to debug Kubernetes pods")
print(f"Embedding dimension: {len(embedding)}")  # 1536
```

### Similarity Search

#### Cosine Similarity
Most common method for comparing embeddings:

```python
import numpy as np
from sklearn.metrics.pairwise import cosine_similarity

def cosine_similarity_search(query_embedding, corpus_embeddings, top_k=5):
    similarities = cosine_similarity([query_embedding], corpus_embeddings)[0]
    top_indices = np.argsort(similarities)[::-1][:top_k]
    return top_indices, similarities[top_indices]

# Example
query = "Kubernetes pod restarting"
query_embedding = model.encode([query])[0]

# Search in corpus
corpus = ["Pod keeps restarting", "Service is down", "Database connection failed"]
corpus_embeddings = model.encode(corpus)

top_indices, scores = cosine_similarity_search(query_embedding, corpus_embeddings)
for idx, score in zip(top_indices, scores):
    print(f"Score: {score:.3f} - {corpus[idx]}")
```

#### Advanced Similarity Metrics

```python
# Euclidean distance
def euclidean_similarity(emb1, emb2):
    return 1 / (1 + np.linalg.norm(emb1 - emb2))

# Dot product (for normalized embeddings)
def dot_product_similarity(emb1, emb2):
    return np.dot(emb1, emb2)

# Manhattan distance
def manhattan_similarity(emb1, emb2):
    return 1 / (1 + np.sum(np.abs(emb1 - emb2)))
```

### Embedding Quality & Evaluation

#### Intrinsic Evaluation
Test embedding quality directly:

```python
def evaluate_embedding_quality(model, test_pairs):
    """Evaluate embedding quality using similarity tasks"""
    correct = 0
    total = 0
    
    for pair1, pair2, expected_similarity in test_pairs:
        emb1 = model.encode([pair1])[0]
        emb2 = model.encode([pair2])[0]
        
        similarity = cosine_similarity([emb1], [emb2])[0][0]
        
        # Check if similarity matches expectation
        if (similarity > 0.7 and expected_similarity == 'high') or \
           (similarity < 0.3 and expected_similarity == 'low'):
            correct += 1
        total += 1
    
    return correct / total

# Example test pairs
test_pairs = [
    ("Kubernetes deployment", "K8s deployment", "high"),
    ("Database error", "Weather forecast", "low"),
    ("Pod restarting", "Container failing", "high")
]

accuracy = evaluate_embedding_quality(model, test_pairs)
print(f"Embedding accuracy: {accuracy:.2%}")
```

#### Extrinsic Evaluation
Test on downstream tasks:

```python
from sklearn.cluster import KMeans
from sklearn.metrics import adjusted_rand_score

def cluster_evaluation(embeddings, true_labels):
    """Evaluate clustering performance"""
    kmeans = KMeans(n_clusters=len(set(true_labels)))
    predicted_labels = kmeans.fit_predict(embeddings)
    
    ari_score = adjusted_rand_score(true_labels, predicted_labels)
    return ari_score
```

### Embedding Optimization

#### Dimensionality Reduction
Reduce embedding dimensions while preserving information:

```python
from sklearn.decomposition import PCA
from sklearn.manifold import TSNE
import umap

# PCA for linear dimensionality reduction
pca = PCA(n_components=128)
reduced_embeddings = pca.fit_transform(embeddings)

# t-SNE for visualization (2D)
tsne = TSNE(n_components=2, random_state=42)
embeddings_2d = tsne.fit_transform(embeddings)

# UMAP for better preservation of local structure
umap_reducer = umap.UMAP(n_components=64, random_state=42)
umap_embeddings = umap_reducer.fit_transform(embeddings)
```

#### Embedding Normalization
Normalize embeddings for consistent similarity computation:

```python
def normalize_embeddings(embeddings):
    """L2 normalize embeddings"""
    norms = np.linalg.norm(embeddings, axis=1, keepdims=True)
    return embeddings / norms

# Normalize before storage
normalized_embeddings = normalize_embeddings(embeddings)
```

### Production Considerations

#### Batch Processing
Process multiple texts efficiently:

```python
def batch_encode_texts(model, texts, batch_size=32):
    """Process texts in batches for efficiency"""
    all_embeddings = []
    
    for i in range(0, len(texts), batch_size):
        batch = texts[i:i + batch_size]
        batch_embeddings = model.encode(batch)
        all_embeddings.extend(batch_embeddings)
    
    return np.array(all_embeddings)
```

#### Caching Embeddings
Cache embeddings to avoid recomputation:

```python
import pickle
import hashlib

class EmbeddingCache:
    def __init__(self, cache_file="embeddings_cache.pkl"):
        self.cache_file = cache_file
        self.cache = self.load_cache()
    
    def load_cache(self):
        try:
            with open(self.cache_file, 'rb') as f:
                return pickle.load(f)
        except FileNotFoundError:
            return {}
    
    def get_embedding(self, text, model):
        text_hash = hashlib.md5(text.encode()).hexdigest()
        
        if text_hash in self.cache:
            return self.cache[text_hash]
        
        embedding = model.encode([text])[0]
        self.cache[text_hash] = embedding
        self.save_cache()
        return embedding
    
    def save_cache(self):
        with open(self.cache_file, 'wb') as f:
            pickle.dump(self.cache, f)
```

**In Agent Bruno**:
- **Semantic Search**: Uses sentence-transformers for document embeddings
- **Memory System**: Stores conversation embeddings in LanceDB
- **Hybrid RAG**: Combines semantic embeddings with keyword search
- **Context Retrieval**: Finds relevant documents using cosine similarity
- **Learning Loop**: Updates embeddings based on user feedback

📖 **Related Docs**: [RAG.md](./RAG.md) | [CONTEXT_CHUNKING.md](./CONTEXT_CHUNKING.md) | [LANCEDB_PERSISTENCE.md](./LANCEDB_PERSISTENCE.md)

---

## 4. Language Models (LLMs)

### What are Language Models?

Language Models are neural networks trained to understand and generate human language. They can perform tasks like text completion, translation, summarization, and question answering.

**Key Capabilities**:
- **Text Generation**: Complete sentences, paragraphs, or entire documents
- **Understanding**: Comprehend context, sentiment, and meaning
- **Reasoning**: Perform logical inference and problem-solving
- **Multilingual**: Work across multiple languages
- **Few-shot Learning**: Adapt to new tasks with minimal examples

### LLM Architecture

#### Transformer Architecture
Modern LLMs are based on the Transformer architecture:

```python
import torch
import torch.nn as nn
from transformers import GPT2LMHeadModel, GPT2Tokenizer

class SimpleLLM(nn.Module):
    def __init__(self, vocab_size, d_model, n_heads, n_layers):
        super().__init__()
        self.embedding = nn.Embedding(vocab_size, d_model)
        self.pos_encoding = nn.Parameter(torch.randn(1000, d_model))
        
        # Transformer blocks
        self.transformer_blocks = nn.ModuleList([
            TransformerBlock(d_model, n_heads) 
            for _ in range(n_layers)
        ])
        
        self.lm_head = nn.Linear(d_model, vocab_size)
    
    def forward(self, input_ids):
        # Embedding + positional encoding
        x = self.embedding(input_ids) + self.pos_encoding[:input_ids.size(1)]
        
        # Pass through transformer blocks
        for block in self.transformer_blocks:
            x = block(x)
        
        # Generate logits for next token prediction
        return self.lm_head(x)
```

#### Key Components

1. **Tokenization**: Convert text to numerical tokens
2. **Embeddings**: Convert tokens to dense vectors
3. **Positional Encoding**: Add position information
4. **Attention Layers**: Focus on relevant parts of input
5. **Feed-forward Networks**: Process information
6. **Output Head**: Generate probability distribution over vocabulary

### Types of Language Models

#### 1. Autoregressive Models (GPT-style)
Generate text one token at a time:

```python
from transformers import GPT2LMHeadModel, GPT2Tokenizer

# Load pre-trained model
model = GPT2LMHeadModel.from_pretrained('gpt2')
tokenizer = GPT2Tokenizer.from_pretrained('gpt2')

def generate_text(prompt, max_length=100, temperature=0.7):
    inputs = tokenizer.encode(prompt, return_tensors='pt')
    
    with torch.no_grad():
        outputs = model.generate(
            inputs,
            max_length=max_length,
            temperature=temperature,
            do_sample=True,
            pad_token_id=tokenizer.eos_token_id
        )
    
    return tokenizer.decode(outputs[0], skip_special_tokens=True)

# Example
prompt = "The future of AI is"
generated = generate_text(prompt)
print(generated)
```

#### 2. Encoder-Decoder Models (T5-style)
Good for tasks like translation and summarization:

```python
from transformers import T5ForConditionalGeneration, T5Tokenizer

model = T5ForConditionalGeneration.from_pretrained('t5-small')
tokenizer = T5Tokenizer.from_pretrained('t5-small')

def summarize_text(text):
    inputs = tokenizer.encode("summarize: " + text, return_tensors='pt')
    
    with torch.no_grad():
        outputs = model.generate(
            inputs,
            max_length=50,
            num_beams=4,
            early_stopping=True
        )
    
    return tokenizer.decode(outputs[0], skip_special_tokens=True)
```

#### 3. Encoder-Only Models (BERT-style)
Excellent for understanding and classification:

```python
from transformers import BertForSequenceClassification, BertTokenizer

model = BertForSequenceClassification.from_pretrained('bert-base-uncased')
tokenizer = BertTokenizer.from_pretrained('bert-base-uncased')

def classify_sentiment(text):
    inputs = tokenizer(text, return_tensors='pt', truncation=True, padding=True)
    
    with torch.no_grad():
        outputs = model(**inputs)
        predictions = torch.nn.functional.softmax(outputs.logits, dim=-1)
    
    return predictions
```

### LLM Training Process

#### Pre-training
Train on large text corpora to learn language patterns:

```python
def pretrain_language_model(model, dataloader, optimizer, num_epochs):
    model.train()
    
    for epoch in range(num_epochs):
        total_loss = 0
        
        for batch in dataloader:
            input_ids = batch['input_ids']
            labels = batch['labels']
            
            # Forward pass
            outputs = model(input_ids=input_ids, labels=labels)
            loss = outputs.loss
            
            # Backward pass
            optimizer.zero_grad()
            loss.backward()
            optimizer.step()
            
            total_loss += loss.item()
        
        print(f"Epoch {epoch}: Loss = {total_loss / len(dataloader):.4f}")
```

#### Fine-tuning
Adapt pre-trained models to specific tasks:

```python
from transformers import Trainer, TrainingArguments

def finetune_model(model, train_dataset, eval_dataset, task_name):
    training_args = TrainingArguments(
        output_dir=f'./results/{task_name}',
        num_train_epochs=3,
        per_device_train_batch_size=16,
        per_device_eval_batch_size=16,
        warmup_steps=500,
        weight_decay=0.01,
        logging_dir=f'./logs/{task_name}',
        logging_steps=10,
        evaluation_strategy="steps",
        eval_steps=500,
        save_strategy="steps",
        save_steps=500,
    )
    
    trainer = Trainer(
        model=model,
        args=training_args,
        train_dataset=train_dataset,
        eval_dataset=eval_dataset,
    )
    
    trainer.train()
    return trainer
```

### Advanced LLM Techniques

#### 1. Prompt Engineering
Design effective prompts to guide model behavior:

```python
def create_few_shot_prompt(examples, query):
    """Create a few-shot prompt with examples"""
    prompt = "Here are some examples:\n\n"
    
    for example in examples:
        prompt += f"Input: {example['input']}\n"
        prompt += f"Output: {example['output']}\n\n"
    
    prompt += f"Input: {query}\nOutput:"
    return prompt

# Example
examples = [
    {"input": "What is Kubernetes?", "output": "Kubernetes is a container orchestration platform."},
    {"input": "How do I deploy a pod?", "output": "Use kubectl apply with a YAML manifest."}
]

query = "What is Docker?"
prompt = create_few_shot_prompt(examples, query)
```

#### 2. Chain-of-Thought Prompting
Encourage step-by-step reasoning:

```python
def cot_prompt(question):
    """Chain-of-thought prompt for complex reasoning"""
    return f"""
    Question: {question}
    
    Let's think step by step:
    1. First, I need to understand what's being asked
    2. Then, I'll break down the problem
    3. Finally, I'll provide a comprehensive answer
    
    Answer:
    """
```

#### 3. Retrieval-Augmented Generation (RAG)
Combine LLMs with external knowledge:

```python
class RAGSystem:
    def __init__(self, llm, retriever, knowledge_base):
        self.llm = llm
        self.retriever = retriever
        self.knowledge_base = knowledge_base
    
    def answer_question(self, question):
        # Retrieve relevant documents
        relevant_docs = self.retriever.search(question, top_k=5)
        
        # Create context
        context = "\n".join([doc.content for doc in relevant_docs])
        
        # Generate answer with context
        prompt = f"""
        Context: {context}
        
        Question: {question}
        
        Answer based on the context:
        """
        
        return self.llm.generate(prompt)
```

### LLM Evaluation

#### Intrinsic Evaluation
Test basic language modeling capabilities:

```python
import math

def evaluate_perplexity(model, tokenizer, test_texts):
    """Calculate perplexity on test data"""
    total_loss = 0
    total_tokens = 0
    
    for text in test_texts:
        inputs = tokenizer(text, return_tensors='pt')
        
        with torch.no_grad():
            outputs = model(**inputs, labels=inputs['input_ids'])
            loss = outputs.loss
            total_loss += loss.item() * inputs['input_ids'].size(1)
            total_tokens += inputs['input_ids'].size(1)
    
    avg_loss = total_loss / total_tokens
    perplexity = math.exp(avg_loss)
    return perplexity
```

#### Extrinsic Evaluation
Test on downstream tasks:

```python
def evaluate_qa_accuracy(model, qa_dataset):
    """Evaluate question-answering accuracy"""
    correct = 0
    total = 0
    
    for item in qa_dataset:
        question = item['question']
        context = item['context']
        true_answer = item['answer']
        
        # Generate answer
        prompt = f"Context: {context}\nQuestion: {question}\nAnswer:"
        generated = model.generate(prompt)
        
        # Simple exact match evaluation
        if true_answer.lower() in generated.lower():
            correct += 1
        total += 1
    
    return correct / total
```

### Production LLM Considerations

#### 1. Model Serving
Deploy LLMs efficiently:

```python
import torch
from transformers import pipeline

class LLMService:
    def __init__(self, model_name, device='cuda'):
        self.device = device
        self.model = pipeline(
            'text-generation',
            model=model_name,
            device=0 if device == 'cuda' else -1
        )
    
    def generate(self, prompt, max_length=100, temperature=0.7):
        return self.model(
            prompt,
            max_length=max_length,
            temperature=temperature,
            do_sample=True,
            pad_token_id=self.model.tokenizer.eos_token_id
        )
```

#### 2. Caching and Optimization
Optimize for production:

```python
import torch
from torch.utils.data import DataLoader

class OptimizedLLM:
    def __init__(self, model, tokenizer):
        self.model = model
        self.tokenizer = tokenizer
        self.model.eval()
        
        # Enable optimizations
        if torch.cuda.is_available():
            self.model = self.model.cuda()
            self.model = torch.compile(self.model)  # PyTorch 2.0 optimization
    
    def batch_generate(self, prompts, batch_size=8):
        """Generate text for multiple prompts efficiently"""
        results = []
        
        for i in range(0, len(prompts), batch_size):
            batch = prompts[i:i + batch_size]
            inputs = self.tokenizer(
                batch, 
                return_tensors='pt', 
                padding=True, 
                truncation=True
            )
            
            if torch.cuda.is_available():
                inputs = {k: v.cuda() for k, v in inputs.items()}
            
            with torch.no_grad():
                outputs = self.model.generate(
                    **inputs,
                    max_length=100,
                    temperature=0.7,
                    do_sample=True
                )
            
            batch_results = self.tokenizer.batch_decode(outputs, skip_special_tokens=True)
            results.extend(batch_results)
        
        return results
```

#### 3. Safety and Alignment
Ensure responsible AI usage:

```python
class SafeLLM:
    def __init__(self, model, safety_classifier):
        self.model = model
        self.safety_classifier = safety_classifier
    
    def safe_generate(self, prompt, max_length=100):
        # Check input safety
        if not self.safety_classifier.is_safe(prompt):
            return "I cannot generate content for this request."
        
        # Generate response
        response = self.model.generate(prompt, max_length=max_length)
        
        # Check output safety
        if not self.safety_classifier.is_safe(response):
            return "I cannot provide this response."
        
        return response
```

**In Agent Bruno**:
- **Ollama Integration**: Uses Ollama for local LLM inference
- **Context-Aware Generation**: Incorporates retrieved documents in responses
- **Memory Integration**: Learns from conversations and updates knowledge
- **Tool Integration**: Can invoke external tools via MCP protocol
- **Safety Measures**: Implements content filtering and safety checks

📖 **Related Docs**: [OLLAMA.md](./OLLAMA.md) | [RAG.md](./RAG.md) | [MEMORY.md](./MEMORY.md)

---

## 5. Retrieval-Augmented Generation (RAG)

### What is RAG?

Retrieval-Augmented Generation (RAG) is a technique that combines the power of large language models with external knowledge retrieval. Instead of relying solely on the model's training data, RAG retrieves relevant information from a knowledge base and uses it to generate more accurate and up-to-date responses.

**Key Benefits**:
- **Access to Current Information**: Can use recent data not in training
- **Reduced Hallucination**: Grounded in retrieved facts
- **Domain-Specific Knowledge**: Can access specialized databases
- **Transparency**: Users can see source documents
- **Cost-Effective**: Reduces need for massive model retraining

### RAG Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        RAG System Architecture                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  User Query                                                     │
│      ↓                                                          │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────────┐   │
│  │   Query     │    │  Retrieval  │    │   Knowledge Base    │   │
│  │ Embedding   │───▶│   Engine    │───▶│   (Vector Store)    │   │
│  └─────────────┘    └─────────────┘    └─────────────────────┘   │
│                           │                                      │
│                           ▼                                      │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │                Retrieved Documents                          │ │
│  └─────────────────────┬───────────────────────────────────────┘ │
│                        │                                        │
│                        ▼                                        │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │              Context + Query                                │ │
│  │                    ↓                                       │ │
│  │              LLM Generation                                │ │
│  │                    ↓                                       │ │
│  │              Final Response                                │ │
│  └─────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

### RAG Components

#### 1. Document Processing Pipeline

```python
import hashlib
from typing import List, Dict
from dataclasses import dataclass

@dataclass
class Document:
    content: str
    metadata: Dict
    embedding: List[float]
    doc_id: str

class DocumentProcessor:
    def __init__(self, chunk_size=512, chunk_overlap=50):
        self.chunk_size = chunk_size
        self.chunk_overlap = chunk_overlap
    
    def chunk_document(self, text: str, metadata: Dict = None) -> List[Document]:
        """Split document into overlapping chunks"""
        chunks = []
        
        for i in range(0, len(text), self.chunk_size - self.chunk_overlap):
            chunk_text = text[i:i + self.chunk_size]
            chunk_id = hashlib.md5(chunk_text.encode()).hexdigest()
            
            chunk = Document(
                content=chunk_text,
                metadata=metadata or {},
                embedding=[],
                doc_id=chunk_id
            )
            chunks.append(chunk)
        
        return chunks
    
    def preprocess_text(self, text: str) -> str:
        """Clean and normalize text"""
        # Remove extra whitespace
        text = ' '.join(text.split())
        
        # Remove special characters (optional)
        # text = re.sub(r'[^\w\s]', '', text)
        
        return text
```

#### 2. Embedding Generation

```python
from sentence_transformers import SentenceTransformer
import numpy as np

class EmbeddingGenerator:
    def __init__(self, model_name='all-MiniLM-L6-v2'):
        self.model = SentenceTransformer(model_name)
    
    def generate_embeddings(self, texts: List[str]) -> np.ndarray:
        """Generate embeddings for a list of texts"""
        return self.model.encode(texts, convert_to_numpy=True)
    
    def generate_single_embedding(self, text: str) -> np.ndarray:
        """Generate embedding for a single text"""
        return self.model.encode([text], convert_to_numpy=True)[0]
```

#### 3. Vector Store Integration

```python
import lancedb
from typing import List, Tuple

class VectorStore:
    def __init__(self, db_path: str):
        self.db = lancedb.connect(db_path)
        self.table = None
    
    def create_table(self, table_name: str, schema):
        """Create a new table for storing vectors"""
        self.table = self.db.create_table(table_name, schema)
    
    def add_documents(self, documents: List[Document]):
        """Add documents to the vector store"""
        if not self.table:
            raise ValueError("Table not created")
        
        # Prepare data for insertion
        data = []
        for doc in documents:
            data.append({
                'id': doc.doc_id,
                'content': doc.content,
                'embedding': doc.embedding,
                'metadata': doc.metadata
            })
        
        self.table.add(data)
    
    def search(self, query_embedding: np.ndarray, top_k: int = 5) -> List[Tuple[Document, float]]:
        """Search for similar documents"""
        if not self.table:
            raise ValueError("Table not created")
        
        results = self.table.search(query_embedding).limit(top_k).to_pandas()
        
        documents = []
        for _, row in results.iterrows():
            doc = Document(
                content=row['content'],
                metadata=row['metadata'],
                embedding=row['embedding'],
                doc_id=row['id']
            )
            documents.append((doc, row['_distance']))
        
        return documents
```

### Hybrid RAG Implementation

#### 1. Semantic + Keyword Search

```python
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.metrics.pairwise import cosine_similarity
import numpy as np

class HybridRetriever:
    def __init__(self, embedding_model, vector_store):
        self.embedding_model = embedding_model
        self.vector_store = vector_store
        self.tfidf_vectorizer = TfidfVectorizer(max_features=1000)
        self.tfidf_matrix = None
        self.documents = []
    
    def build_tfidf_index(self, documents: List[Document]):
        """Build TF-IDF index for keyword search"""
        self.documents = documents
        texts = [doc.content for doc in documents]
        self.tfidf_matrix = self.tfidf_vectorizer.fit_transform(texts)
    
    def semantic_search(self, query: str, top_k: int = 5) -> List[Tuple[Document, float]]:
        """Perform semantic search using embeddings"""
        query_embedding = self.embedding_model.generate_single_embedding(query)
        return self.vector_store.search(query_embedding, top_k)
    
    def keyword_search(self, query: str, top_k: int = 5) -> List[Tuple[Document, float]]:
        """Perform keyword search using BM25 (preferred) or TF-IDF"""
        if self.tfidf_matrix is None:
            raise ValueError("TF-IDF index not built")
        
        query_vector = self.tfidf_vectorizer.transform([query])
        similarities = cosine_similarity(query_vector, self.tfidf_matrix)[0]
        
        # Get top-k results
        top_indices = np.argsort(similarities)[::-1][:top_k]
        
        results = []
        for idx in top_indices:
            if similarities[idx] > 0:  # Only include non-zero similarities
                results.append((self.documents[idx], similarities[idx]))
        
        return results
    
    def hybrid_search(self, query: str, top_k: int = 5, alpha: float = 0.7) -> List[Tuple[Document, float]]:
        """Combine semantic and keyword search"""
        # Get results from both methods
        semantic_results = self.semantic_search(query, top_k * 2)
        keyword_results = self.keyword_search(query, top_k * 2)
        
        # Create score dictionary
        doc_scores = {}
        
        # Add semantic scores
        for doc, score in semantic_results:
            doc_scores[doc.doc_id] = doc_scores.get(doc.doc_id, 0) + alpha * score
        
        # Add keyword scores
        for doc, score in keyword_results:
            doc_scores[doc.doc_id] = doc_scores.get(doc.doc_id, 0) + (1 - alpha) * score
        
        # Sort by combined score
        sorted_results = sorted(doc_scores.items(), key=lambda x: x[1], reverse=True)
        
        # Return top-k results
        final_results = []
        for doc_id, score in sorted_results[:top_k]:
            # Find the document
            for doc, _ in semantic_results + keyword_results:
                if doc.doc_id == doc_id:
                    final_results.append((doc, score))
                    break
        
        return final_results
```

#### 2. Re-ranking with Cross-Encoders

```python
from sentence_transformers import CrossEncoder

class Reranker:
    def __init__(self, model_name='cross-encoder/ms-marco-MiniLM-L-6-v2'):
        self.model = CrossEncoder(model_name)
    
    def rerank(self, query: str, documents: List[Document], top_k: int = 5) -> List[Tuple[Document, float]]:
        """Re-rank documents using cross-encoder"""
        # Prepare query-document pairs
        pairs = [(query, doc.content) for doc in documents]
        
        # Get relevance scores
        scores = self.model.predict(pairs)
        
        # Combine documents with scores
        doc_scores = list(zip(documents, scores))
        
        # Sort by score (descending)
        doc_scores.sort(key=lambda x: x[1], reverse=True)
        
        return doc_scores[:top_k]
```

### BM25: The Gold Standard for Keyword Search

**BM25 (Best Matching 25)** is the industry-standard probabilistic ranking function for keyword-based search. It's significantly more effective than TF-IDF and is used by search engines worldwide, including Elasticsearch, Lucene, and Agent Bruno's hybrid RAG system.

#### Why BM25 is Superior to TF-IDF

**TF-IDF Problems**:
- Doesn't consider document length properly
- Score grows unbounded with term frequency
- No tuning parameters for different use cases

**BM25 Solutions**:
- Saturation function prevents term frequency inflation
- Document length normalization
- Tunable parameters (k₁, b) for optimization
- Better aligns with relevance judgments

#### BM25 Formula

```
BM25(D, Q) = Σ IDF(qᵢ) · (f(qᵢ, D) · (k₁ + 1)) / (f(qᵢ, D) + k₁ · (1 - b + b · |D| / avgDL))
          qᵢ∈Q

Where:
- Q = query tokens
- D = document
- f(qᵢ, D) = frequency of term qᵢ in document D
- |D| = length of document D (in words)
- avgDL = average document length in collection
- k₁ = term frequency saturation parameter (default: 1.5)
  * Controls how quickly term frequency contribution saturates
  * Higher k₁ → more weight to term frequency
  * Lower k₁ → quicker saturation
- b = document length normalization parameter (default: 0.75)
  * Controls how much document length affects ranking
  * b=0 → no normalization
  * b=1 → full normalization
  * 0 < b < 1 → partial normalization
- IDF(qᵢ) = inverse document frequency of term qᵢ
  * IDF(qᵢ) = log((N - df(qᵢ) + 0.5) / (df(qᵢ) + 0.5) + 1)
  * N = total number of documents
  * df(qᵢ) = number of documents containing term qᵢ
```

#### BM25 Implementation for Agent Bruno

```python
import numpy as np
from typing import List, Dict, Tuple
from collections import Counter
import math
from dataclasses import dataclass

@dataclass
class BM25Document:
    """Document representation for BM25"""
    doc_id: str
    content: str
    tokens: List[str]
    token_freqs: Dict[str, int]
    length: int  # Number of tokens

class BM25Retriever:
    """
    Production-ready BM25 implementation for Agent Bruno.
    
    Used in hybrid RAG alongside semantic (embedding) search.
    """
    
    def __init__(self, k1: float = 1.5, b: float = 0.75):
        """
        Initialize BM25 retriever.
        
        Args:
            k1: Term frequency saturation parameter (1.2-2.0 typical)
                - Lower k1 (0.5-1.0): Better for short queries
                - Higher k1 (2.0-3.0): Better for long documents
            b: Document length normalization (0.0-1.0)
               - b=0.75: Standard (recommended for most cases)
               - b=0.0: No normalization (good for similar-length docs)
               - b=1.0: Full normalization (good for varied-length docs)
        """
        self.k1 = k1
        self.b = b
        self.documents: List[BM25Document] = []
        self.doc_freqs: Dict[str, int] = {}  # Document frequency per term
        self.idf: Dict[str, float] = {}  # IDF score per term
        self.avgdl: float = 0.0  # Average document length
        self.N: int = 0  # Total number of documents
    
    def tokenize(self, text: str) -> List[str]:
        """
        Tokenize text for BM25.
        
        Steps:
        1. Lowercase
        2. Remove punctuation
        3. Split on whitespace
        4. Remove stopwords (optional)
        5. Stemming (optional)
        """
        import re
        from nltk.corpus import stopwords
        from nltk.stem import PorterStemmer
        
        # Lowercase and remove punctuation
        text = re.sub(r'[^\w\s]', ' ', text.lower())
        
        # Split into tokens
        tokens = text.split()
        
        # Remove stopwords (common words like "the", "is", "at")
        stop_words = set(stopwords.words('english'))
        tokens = [t for t in tokens if t not in stop_words and len(t) > 2]
        
        # Stemming (reduce words to root form)
        # "running" → "run", "better" → "better"
        stemmer = PorterStemmer()
        tokens = [stemmer.stem(t) for t in tokens]
        
        return tokens
    
    def build_index(self, documents: List[Dict[str, str]]):
        """
        Build BM25 index from documents.
        
        Args:
            documents: List of dicts with 'id' and 'content' keys
        """
        self.documents = []
        self.N = len(documents)
        
        # Process all documents
        for doc in documents:
            tokens = self.tokenize(doc['content'])
            token_freqs = Counter(tokens)
            
            bm25_doc = BM25Document(
                doc_id=doc['id'],
                content=doc['content'],
                tokens=tokens,
                token_freqs=token_freqs,
                length=len(tokens)
            )
            self.documents.append(bm25_doc)
        
        # Calculate average document length
        self.avgdl = sum(doc.length for doc in self.documents) / self.N
        
        # Calculate document frequencies
        self.doc_freqs = {}
        for doc in self.documents:
            for token in set(doc.tokens):  # Unique tokens only
                self.doc_freqs[token] = self.doc_freqs.get(token, 0) + 1
        
        # Calculate IDF for each term
        self.idf = {}
        for term, df in self.doc_freqs.items():
            # BM25 IDF formula (with +1 to avoid log(0))
            idf_score = math.log((self.N - df + 0.5) / (df + 0.5) + 1.0)
            self.idf[term] = idf_score
        
        print(f"BM25 index built: {self.N} documents, {len(self.doc_freqs)} unique terms")
        print(f"Average document length: {self.avgdl:.2f} tokens")
    
    def score_document(self, query_tokens: List[str], doc: BM25Document) -> float:
        """
        Calculate BM25 score for a single document.
        
        Returns:
            BM25 score (higher = more relevant)
        """
        score = 0.0
        
        for token in query_tokens:
            if token not in doc.token_freqs:
                continue  # Term not in document
            
            # Term frequency in this document
            f = doc.token_freqs[token]
            
            # IDF score for this term
            idf = self.idf.get(token, 0.0)
            
            # Document length normalization factor
            norm = 1 - self.b + self.b * (doc.length / self.avgdl)
            
            # BM25 term score
            term_score = idf * (f * (self.k1 + 1)) / (f + self.k1 * norm)
            
            score += term_score
        
        return score
    
    def search(self, query: str, top_k: int = 10) -> List[Tuple[str, float, str]]:
        """
        Search documents using BM25.
        
        Args:
            query: Search query
            top_k: Number of results to return
        
        Returns:
            List of (doc_id, score, content) tuples, sorted by relevance
        """
        # Tokenize query
        query_tokens = self.tokenize(query)
        
        if not query_tokens:
            return []
        
        # Score all documents
        doc_scores = []
        for doc in self.documents:
            score = self.score_document(query_tokens, doc)
            if score > 0:  # Only include docs with non-zero scores
                doc_scores.append((doc.doc_id, score, doc.content))
        
        # Sort by score (descending)
        doc_scores.sort(key=lambda x: x[1], reverse=True)
        
        return doc_scores[:top_k]
    
    def get_term_stats(self, term: str) -> Dict[str, float]:
        """
        Get statistics for a term (useful for debugging).
        
        Returns:
            Dictionary with term frequency, document frequency, and IDF
        """
        tokens = self.tokenize(term)
        if not tokens:
            return {}
        
        token = tokens[0]
        df = self.doc_freqs.get(token, 0)
        idf = self.idf.get(token, 0.0)
        
        return {
            'term': token,
            'document_frequency': df,
            'idf': idf,
            'appears_in_pct': (df / self.N * 100) if self.N > 0 else 0.0
        }

# Usage example for Agent Bruno
if __name__ == "__main__":
    # Sample documents from Agent Bruno's knowledge base
    documents = [
        {
            'id': 'doc1',
            'content': 'Loki crashes are commonly caused by out of memory errors. Check pod logs and increase memory limits.'
        },
        {
            'id': 'doc2',
            'content': 'Grafana dashboard shows Loki ingestion rate is too high. Scale up Loki ingesters to handle load.'
        },
        {
            'id': 'doc3',
            'content': 'Prometheus alerting rules for Loki include high crash rate and memory usage warnings.'
        },
        {
            'id': 'doc4',
            'content': 'Kubernetes pod restart policy affects how Loki containers behave after crashes.'
        },
        {
            'id': 'doc5',
            'content': 'Memory limits and requests should be configured properly for Loki deployment.'
        }
    ]
    
    # Build BM25 index
    retriever = BM25Retriever(k1=1.5, b=0.75)
    retriever.build_index(documents)
    
    # Search query
    query = "Loki crashes memory"
    print(f"\nQuery: '{query}'")
    print("=" * 80)
    
    # Get results
    results = retriever.search(query, top_k=3)
    
    for i, (doc_id, score, content) in enumerate(results, 1):
        print(f"\nRank {i}: {doc_id} (score: {score:.4f})")
        print(f"Content: {content[:100]}...")
    
    # Term statistics
    print("\n" + "=" * 80)
    print("Term Statistics:")
    for term in ['loki', 'crashes', 'memory']:
        stats = retriever.get_term_stats(term)
        print(f"\n'{term}':")
        print(f"  Document Frequency: {stats['document_frequency']}")
        print(f"  IDF Score: {stats['idf']:.4f}")
        print(f"  Appears in: {stats['appears_in_pct']:.1f}% of documents")
```

#### BM25 Parameter Tuning

Finding optimal k₁ and b values for your specific use case:

```python
from sklearn.metrics import ndcg_score
import numpy as np

class BM25Tuner:
    """
    Tune BM25 parameters using grid search.
    
    Requires labeled relevance judgments for a set of queries.
    """
    
    def __init__(self, documents: List[Dict], queries_with_labels: List[Dict]):
        """
        Args:
            documents: List of documents to index
            queries_with_labels: List of dicts with 'query' and 'relevant_doc_ids'
        """
        self.documents = documents
        self.queries = queries_with_labels
    
    def evaluate(self, k1: float, b: float) -> float:
        """
        Evaluate BM25 with given parameters.
        
        Returns:
            Mean NDCG@10 score across all queries
        """
        retriever = BM25Retriever(k1=k1, b=b)
        retriever.build_index(self.documents)
        
        ndcg_scores = []
        
        for query_data in self.queries:
            query = query_data['query']
            relevant_doc_ids = set(query_data['relevant_doc_ids'])
            
            # Get search results
            results = retriever.search(query, top_k=10)
            
            # Create relevance labels (1 if relevant, 0 if not)
            y_true = [1 if doc_id in relevant_doc_ids else 0 
                      for doc_id, _, _ in results]
            
            # Scores from BM25
            y_score = [score for _, score, _ in results]
            
            if sum(y_true) > 0:  # Only if there are relevant docs
                ndcg = ndcg_score([y_true], [y_score])
                ndcg_scores.append(ndcg)
        
        return np.mean(ndcg_scores) if ndcg_scores else 0.0
    
    def grid_search(self) -> Tuple[float, float, float]:
        """
        Grid search over k1 and b parameters.
        
        Returns:
            (best_k1, best_b, best_score)
        """
        k1_values = [0.5, 1.0, 1.2, 1.5, 1.8, 2.0]
        b_values = [0.0, 0.25, 0.5, 0.75, 1.0]
        
        best_k1 = 1.5
        best_b = 0.75
        best_score = 0.0
        
        print("Tuning BM25 parameters...")
        print(f"{'k1':>6s} {'b':>6s} {'NDCG@10':>10s}")
        print("-" * 25)
        
        for k1 in k1_values:
            for b in b_values:
                score = self.evaluate(k1, b)
                print(f"{k1:>6.2f} {b:>6.2f} {score:>10.4f}")
                
                if score > best_score:
                    best_score = score
                    best_k1 = k1
                    best_b = b
        
        print("-" * 25)
        print(f"Best parameters: k1={best_k1:.2f}, b={best_b:.2f}, score={best_score:.4f}")
        
        return best_k1, best_b, best_score

# Usage for Agent Bruno
# Tune on a validation set of queries with known relevant documents
tuner = BM25Tuner(documents, validation_queries)
best_k1, best_b, score = tuner.grid_search()

# Use optimized parameters in production
production_retriever = BM25Retriever(k1=best_k1, b=best_b)
production_retriever.build_index(documents)
```

#### Integrating BM25 with Agent Bruno's Hybrid RAG

```python
class HybridRAGWithBM25:
    """
    Production hybrid RAG system combining:
    - Semantic search (dense vectors via nomic-embed-text)
    - Keyword search (BM25)
    - Re-ranking (cross-encoder)
    """
    
    def __init__(self, embedding_model, vector_store, bm25_retriever):
        self.embedding_model = embedding_model
        self.vector_store = vector_store
        self.bm25 = bm25_retriever
    
    def retrieve(
        self,
        query: str,
        top_k: int = 5,
        semantic_weight: float = 0.6,
        bm25_weight: float = 0.4
    ) -> List[Tuple[str, float, str]]:
        """
        Hybrid retrieval with RRF (Reciprocal Rank Fusion).
        
        Args:
            query: User query
            top_k: Number of results to return
            semantic_weight: Weight for semantic search (0-1)
            bm25_weight: Weight for BM25 (should sum to 1 with semantic_weight)
        
        Returns:
            List of (doc_id, fused_score, content) tuples
        """
        # 1. Semantic search (embedding similarity)
        query_embedding = self.embedding_model.generate_single_embedding(query)
        semantic_results = self.vector_store.search(query_embedding, top_k=20)
        
        # 2. BM25 keyword search
        bm25_results = self.bm25.search(query, top_k=20)
        
        # 3. Reciprocal Rank Fusion (RRF)
        # Combines rankings from both methods
        doc_scores = {}
        k = 60  # RRF constant (tunable)
        
        # Add semantic scores
        for rank, (doc_id, score, content) in enumerate(semantic_results, 1):
            rrf_score = semantic_weight / (k + rank)
            doc_scores[doc_id] = doc_scores.get(doc_id, 0) + rrf_score
        
        # Add BM25 scores
        for rank, (doc_id, score, content) in enumerate(bm25_results, 1):
            rrf_score = bm25_weight / (k + rank)
            doc_scores[doc_id] = doc_scores.get(doc_id, 0) + rrf_score
        
        # Sort by fused score
        sorted_docs = sorted(doc_scores.items(), key=lambda x: x[1], reverse=True)
        
        # Get document content
        doc_map = {doc_id: (score, content) 
                   for doc_id, score, content in semantic_results + bm25_results}
        
        results = []
        for doc_id, fused_score in sorted_docs[:top_k]:
            if doc_id in doc_map:
                _, content = doc_map[doc_id]
                results.append((doc_id, fused_score, content))
        
        return results

# Usage in Agent Bruno
hybrid_rag = HybridRAGWithBM25(
    embedding_model=embedding_generator,
    vector_store=lancedb_store,
    bm25_retriever=bm25_retriever
)

query = "How do I fix Loki crashes?"
results = hybrid_rag.retrieve(query, top_k=5)

for i, (doc_id, score, content) in enumerate(results, 1):
    print(f"{i}. {doc_id} (score: {score:.4f})")
    print(f"   {content[:100]}...\n")
```

**Key Insights for Agent Bruno**:
- **BM25 complements semantic search**: Catches exact keyword matches that embeddings might miss
- **Optimal for technical queries**: Keywords like "Loki", "Kubernetes", "crash" are important exact matches
- **Parameter recommendations for SRE docs**:
  - k₁=1.2: Technical docs have shorter, precise text
  - b=0.75: Standard normalization works well
- **RRF fusion**: Combines best of both worlds without complex score calibration

📖 **Related Docs**: [RAG.md](./RAG.md) | [FUSION_RE_RANKING.md](./FUSION_RE_RANKING.md) | [LANCEDB_PERSISTENCE.md](./LANCEDB_PERSISTENCE.md)

---

### Complete RAG System

```python
class RAGSystem:
    def __init__(self, embedding_model_name='all-MiniLM-L6-v2', db_path='./rag_db'):
        self.embedding_generator = EmbeddingGenerator(embedding_model_name)
        self.vector_store = VectorStore(db_path)
        self.retriever = HybridRetriever(self.embedding_generator, self.vector_store)
        self.reranker = Reranker()
        self.llm = None  # Will be set separately
    
    def add_documents(self, documents: List[Document]):
        """Add documents to the knowledge base"""
        # Generate embeddings
        texts = [doc.content for doc in documents]
        embeddings = self.embedding_generator.generate_embeddings(texts)
        
        # Add embeddings to documents
        for doc, embedding in zip(documents, embeddings):
            doc.embedding = embedding.tolist()
        
        # Add to vector store
        self.vector_store.add_documents(documents)
        
        # Build TF-IDF index
        self.retriever.build_tfidf_index(documents)
    
    def retrieve(self, query: str, top_k: int = 10) -> List[Document]:
        """Retrieve relevant documents for a query"""
        # Hybrid search
        results = self.retriever.hybrid_search(query, top_k * 2)
        
        # Extract documents
        documents = [doc for doc, score in results]
        
        # Re-rank
        reranked = self.reranker.rerank(query, documents, top_k)
        
        return [doc for doc, score in reranked]
    
    def generate_response(self, query: str, context_docs: List[Document]) -> str:
        """Generate response using retrieved context"""
        # Create context from retrieved documents
        context = "\n\n".join([doc.content for doc in context_docs])
        
        # Create prompt
        prompt = f"""
        Context:
        {context}
        
        Question: {query}
        
        Answer based on the context above:
        """
        
        # Generate response (this would use your LLM)
        # response = self.llm.generate(prompt)
        # return response
        
        # Placeholder
        return f"Based on the context, here's what I found: {len(context_docs)} relevant documents."
```

### RAG Evaluation

#### 1. Retrieval Quality Metrics

```python
def evaluate_retrieval_quality(retriever, test_queries, ground_truth):
    """Evaluate retrieval quality using standard metrics"""
    results = {
        'precision_at_5': [],
        'recall_at_5': [],
        'mrr': []  # Mean Reciprocal Rank
    }
    
    for query, relevant_docs in zip(test_queries, ground_truth):
        # Retrieve documents
        retrieved = retriever.retrieve(query, top_k=5)
        retrieved_ids = {doc.doc_id for doc in retrieved}
        relevant_ids = {doc.doc_id for doc in relevant_docs}
        
        # Calculate metrics
        if retrieved_ids:
            precision = len(retrieved_ids & relevant_ids) / len(retrieved_ids)
            results['precision_at_5'].append(precision)
        
        if relevant_ids:
            recall = len(retrieved_ids & relevant_ids) / len(relevant_ids)
            results['recall_at_5'].append(recall)
        
        # MRR calculation
        for i, doc in enumerate(retrieved):
            if doc.doc_id in relevant_ids:
                results['mrr'].append(1 / (i + 1))
                break
        else:
            results['mrr'].append(0)
    
    # Calculate averages
    return {
        metric: np.mean(scores) 
        for metric, scores in results.items()
    }
```

#### 2. End-to-End RAG Evaluation

```python
def evaluate_rag_system(rag_system, test_cases):
    """Evaluate complete RAG system"""
    results = {
        'retrieval_accuracy': [],
        'response_quality': [],
        'response_relevance': []
    }
    
    for test_case in test_cases:
        query = test_case['query']
        expected_docs = test_case['relevant_docs']
        expected_answer = test_case['expected_answer']
        
        # Test retrieval
        retrieved_docs = rag_system.retrieve(query, top_k=5)
        retrieved_ids = {doc.doc_id for doc in retrieved_docs}
        expected_ids = {doc.doc_id for doc in expected_docs}
        
        retrieval_accuracy = len(retrieved_ids & expected_ids) / len(expected_ids)
        results['retrieval_accuracy'].append(retrieval_accuracy)
        
        # Test response generation
        response = rag_system.generate_response(query, retrieved_docs)
        
        # Simple quality metrics (in practice, use more sophisticated evaluation)
        response_quality = len(response.split()) / 100  # Normalize by length
        results['response_quality'].append(response_quality)
        
        # Relevance (simplified)
        relevance_score = 0.8 if any(keyword in response.lower() for keyword in query.lower().split()) else 0.3
        results['response_relevance'].append(relevance_score)
    
    return {
        metric: np.mean(scores) 
        for metric, scores in results.items()
    }
```

### Advanced RAG Techniques

#### 1. Multi-Query RAG

```python
class MultiQueryRAG:
    def __init__(self, rag_system, query_generator):
        self.rag_system = rag_system
        self.query_generator = query_generator
    
    def retrieve_with_multiple_queries(self, original_query: str, num_queries: int = 3):
        """Generate multiple queries and retrieve documents for each"""
        # Generate alternative queries
        alternative_queries = self.query_generator.generate_queries(original_query, num_queries)
        
        # Retrieve documents for each query
        all_documents = []
        for query in [original_query] + alternative_queries:
            docs = self.rag_system.retrieve(query, top_k=5)
            all_documents.extend(docs)
        
        # Remove duplicates and re-rank
        unique_docs = list({doc.doc_id: doc for doc in all_documents}.values())
        return self.rag_system.reranker.rerank(original_query, unique_docs, top_k=10)
```

#### 2. Adaptive RAG

```python
class AdaptiveRAG:
    def __init__(self, rag_system, confidence_threshold=0.7):
        self.rag_system = rag_system
        self.confidence_threshold = confidence_threshold
    
    def adaptive_retrieve(self, query: str):
        """Adapt retrieval strategy based on query characteristics"""
        # Analyze query complexity
        query_length = len(query.split())
        
        if query_length < 3:
            # Simple query - use semantic search only
            return self.rag_system.retriever.semantic_search(query, top_k=5)
        elif query_length > 10:
            # Complex query - use hybrid search with more documents
            return self.rag_system.retriever.hybrid_search(query, top_k=15)
        else:
            # Medium complexity - use standard hybrid search
            return self.rag_system.retriever.hybrid_search(query, top_k=10)
```

**In Agent Bruno**:
- **Hybrid Retrieval**: Combines semantic and keyword search for better results
- **Context Chunking**: Splits documents into optimal chunks for retrieval
- **Re-ranking**: Uses cross-encoders to improve document ranking
- **Memory Integration**: Stores conversation context for better retrieval
- **Learning Loop**: Updates retrieval based on user feedback
- **MCP Integration**: Can retrieve from external knowledge sources

📖 **Related Docs**: [RAG.md](./RAG.md) | [FUSION_RE_RANKING.md](./FUSION_RE_RANKING.md) | [CONTEXT_CHUNKING.md](./CONTEXT_CHUNKING.md)

---

## 6. Fine-tuning & LoRA

### What is Fine-tuning?

Fine-tuning is the process of adapting a pre-trained model to a specific task or domain by training it on task-specific data. This allows you to leverage the general knowledge learned during pre-training while specializing the model for your use case.

**Benefits of Fine-tuning**:
- **Task Specialization**: Adapt models to specific domains (medical, legal, technical)
- **Improved Performance**: Better results than zero-shot or few-shot prompting
- **Cost Efficiency**: More cost-effective than training from scratch
- **Data Efficiency**: Requires less data than pre-training
- **Customization**: Tailor model behavior to specific requirements

### Types of Fine-tuning

#### 1. Full Fine-tuning
Update all model parameters:

```python
import torch
from transformers import AutoModelForCausalLM, AutoTokenizer, TrainingArguments, Trainer

def full_finetune(model_name, train_dataset, eval_dataset):
    """Full fine-tuning of a language model"""
    # Load model and tokenizer
    model = AutoModelForCausalLM.from_pretrained(model_name)
    tokenizer = AutoTokenizer.from_pretrained(model_name)
    
    # Set pad token
    if tokenizer.pad_token is None:
        tokenizer.pad_token = tokenizer.eos_token
    
    # Training arguments
    training_args = TrainingArguments(
        output_dir='./fine-tuned-model',
        num_train_epochs=3,
        per_device_train_batch_size=4,
        per_device_eval_batch_size=4,
        warmup_steps=100,
        weight_decay=0.01,
        logging_dir='./logs',
        logging_steps=10,
        evaluation_strategy="steps",
        eval_steps=500,
        save_strategy="steps",
        save_steps=500,
        load_best_model_at_end=True,
        metric_for_best_model="eval_loss",
        greater_is_better=False,
    )
    
    # Create trainer
    trainer = Trainer(
        model=model,
        args=training_args,
        train_dataset=train_dataset,
        eval_dataset=eval_dataset,
        tokenizer=tokenizer,
    )
    
    # Train
    trainer.train()
    
    return trainer, model, tokenizer
```

#### 2. Parameter-Efficient Fine-tuning (PEFT)

PEFT methods update only a small subset of parameters, making fine-tuning more efficient:

```python
from peft import LoraConfig, get_peft_model, TaskType
from transformers import AutoModelForCausalLM

def setup_lora(model_name, target_modules=None):
    """Setup LoRA (Low-Rank Adaptation) for efficient fine-tuning"""
    # Load base model
    model = AutoModelForCausalLM.from_pretrained(
        model_name,
        torch_dtype=torch.float16,
        device_map="auto"
    )
    
    # LoRA configuration
    lora_config = LoraConfig(
        task_type=TaskType.CAUSAL_LM,
        inference_mode=False,
        r=16,  # Rank of adaptation
        lora_alpha=32,  # LoRA scaling parameter
        lora_dropout=0.1,
        target_modules=target_modules or ["q_proj", "v_proj", "k_proj", "o_proj"]
    )
    
    # Apply LoRA to model
    model = get_peft_model(model, lora_config)
    
    # Print trainable parameters
    model.print_trainable_parameters()
    
    return model

def train_lora_model(model, train_dataset, eval_dataset):
    """Train a LoRA-adapted model"""
    training_args = TrainingArguments(
        output_dir='./lora-model',
        num_train_epochs=3,
        per_device_train_batch_size=8,  # Can use larger batch size with LoRA
        per_device_eval_batch_size=8,
        warmup_steps=100,
        weight_decay=0.01,
        logging_dir='./logs',
        logging_steps=10,
        evaluation_strategy="steps",
        eval_steps=500,
        save_strategy="steps",
        save_steps=500,
        load_best_model_at_end=True,
    )
    
    trainer = Trainer(
        model=model,
        args=training_args,
        train_dataset=train_dataset,
        eval_dataset=eval_dataset,
    )
    
    trainer.train()
    return trainer
```

### Advanced PEFT Methods

#### 1. QLoRA (Quantized LoRA)
Combine quantization with LoRA for even more efficiency:

```python
from transformers import BitsAndBytesConfig
import torch

def setup_qlora(model_name):
    """Setup QLoRA with 4-bit quantization"""
    # 4-bit quantization configuration
    bnb_config = BitsAndBytesConfig(
        load_in_4bit=True,
        bnb_4bit_use_double_quant=True,
        bnb_4bit_quant_type="nf4",
        bnb_4bit_compute_dtype=torch.bfloat16
    )
    
    # Load model with quantization
    model = AutoModelForCausalLM.from_pretrained(
        model_name,
        quantization_config=bnb_config,
        device_map="auto"
    )
    
    # LoRA configuration for QLoRA
    lora_config = LoraConfig(
        task_type=TaskType.CAUSAL_LM,
        r=16,
        lora_alpha=32,
        lora_dropout=0.1,
        target_modules=["q_proj", "v_proj", "k_proj", "o_proj", "gate_proj", "up_proj", "down_proj"]
    )
    
    model = get_peft_model(model, lora_config)
    return model
```

#### 2. AdaLoRA (Adaptive LoRA)
Dynamically adjust LoRA rank during training:

```python
from peft import AdaLoraConfig, get_peft_model

def setup_adalora(model_name):
    """Setup AdaLoRA for adaptive rank adjustment"""
    model = AutoModelForCausalLM.from_pretrained(model_name)
    
    # AdaLoRA configuration
    adalora_config = AdaLoraConfig(
        task_type=TaskType.CAUSAL_LM,
        init_r=12,
        target_r=8,
        beta1=0.85,
        beta2=0.85,
        tinit=200,
        tfinal=1000,
        deltaT=10,
        lora_alpha=32,
        lora_dropout=0.1,
        target_modules=["q_proj", "v_proj", "k_proj", "o_proj"]
    )
    
    model = get_peft_model(model, adalora_config)
    return model
```

#### 3. Prefix Tuning
Add learnable prefix tokens to the input:

```python
from peft import PrefixTuningConfig, get_peft_model

def setup_prefix_tuning(model_name):
    """Setup prefix tuning for efficient fine-tuning"""
    model = AutoModelForCausalLM.from_pretrained(model_name)
    
    # Prefix tuning configuration
    prefix_config = PrefixTuningConfig(
        task_type=TaskType.CAUSAL_LM,
        num_virtual_tokens=20,
        prefix_projection=True
    )
    
    model = get_peft_model(model, prefix_config)
    return model
```

### Fine-tuning Best Practices

#### 1. Data Preparation

```python
from datasets import Dataset
import json

def prepare_finetuning_data(data_path, tokenizer, max_length=512):
    """Prepare data for fine-tuning"""
    with open(data_path, 'r') as f:
        data = json.load(f)
    
    def tokenize_function(examples):
        # Combine input and target
        texts = [f"{inp} {target}" for inp, target in zip(examples['input'], examples['target'])]
        
        # Tokenize
        tokenized = tokenizer(
            texts,
            truncation=True,
            padding=True,
            max_length=max_length,
            return_tensors="pt"
        )
        
        # Create labels (same as input_ids for causal LM)
        tokenized['labels'] = tokenized['input_ids'].clone()
        
        return tokenized
    
    # Create dataset
    dataset = Dataset.from_dict({
        'input': [item['input'] for item in data],
        'target': [item['target'] for item in data]
    })
    
    # Tokenize
    tokenized_dataset = dataset.map(
        tokenize_function,
        batched=True,
        remove_columns=dataset.column_names
    )
    
    return tokenized_dataset
```

#### 2. Learning Rate Scheduling

```python
from transformers import get_linear_schedule_with_warmup
import torch

def setup_optimizer_and_scheduler(model, num_training_steps, learning_rate=2e-5):
    """Setup optimizer and learning rate scheduler"""
    # Optimizer
    optimizer = torch.optim.AdamW(
        model.parameters(),
        lr=learning_rate,
        weight_decay=0.01
    )
    
    # Learning rate scheduler
    scheduler = get_linear_schedule_with_warmup(
        optimizer,
        num_warmup_steps=num_training_steps * 0.1,  # 10% warmup
        num_training_steps=num_training_steps
    )
    
    return optimizer, scheduler
```

#### 3. Evaluation and Monitoring

```python
import torch
from torch.utils.data import DataLoader
from tqdm import tqdm

def evaluate_model(model, eval_dataloader, device):
    """Evaluate model on validation set"""
    model.eval()
    total_loss = 0
    num_batches = 0
    
    with torch.no_grad():
        for batch in tqdm(eval_dataloader, desc="Evaluating"):
            batch = {k: v.to(device) for k, v in batch.items()}
            
            outputs = model(**batch)
            loss = outputs.loss
            total_loss += loss.item()
            num_batches += 1
    
    avg_loss = total_loss / num_batches
    perplexity = torch.exp(torch.tensor(avg_loss))
    
    return {
        'eval_loss': avg_loss,
        'perplexity': perplexity.item()
    }
```

### Domain-Specific Fine-tuning

#### 1. Technical Documentation

```python
def create_technical_dataset():
    """Create dataset for technical documentation fine-tuning"""
    technical_examples = [
        {
            "input": "How to deploy a Kubernetes cluster?",
            "target": "To deploy a Kubernetes cluster, you can use tools like kubeadm, kops, or managed services like EKS, GKE, or AKS. For a local development cluster, you can use minikube or kind."
        },
        {
            "input": "What is Docker?",
            "target": "Docker is a containerization platform that allows you to package applications and their dependencies into lightweight, portable containers. It uses containerization technology to ensure consistent deployment across different environments."
        },
        {
            "input": "How to debug a failing pod?",
            "target": "To debug a failing pod, use 'kubectl describe pod <pod-name>' to see events and status, 'kubectl logs <pod-name>' to check logs, and 'kubectl exec -it <pod-name> -- /bin/bash' to access the container for debugging."
        }
    ]
    
    return technical_examples
```

#### 2. Code Generation

```python
def create_code_dataset():
    """Create dataset for code generation fine-tuning"""
    code_examples = [
        {
            "input": "Write a Python function to calculate fibonacci numbers:",
            "target": "def fibonacci(n):\n    if n <= 1:\n        return n\n    return fibonacci(n-1) + fibonacci(n-2)"
        },
        {
            "input": "Create a Kubernetes deployment YAML:",
            "target": "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: my-app\nspec:\n  replicas: 3\n  selector:\n    matchLabels:\n      app: my-app\n  template:\n    metadata:\n      labels:\n        app: my-app\n    spec:\n      containers:\n      - name: my-app\n        image: my-app:latest\n        ports:\n        - containerPort: 8080"
        }
    ]
    
    return code_examples
```

### Fine-tuning for Specific Tasks

#### 1. Question Answering

```python
def setup_qa_finetuning(model_name, qa_dataset):
    """Setup fine-tuning for question answering"""
    model = AutoModelForCausalLM.from_pretrained(model_name)
    tokenizer = AutoTokenizer.from_pretrained(model_name)
    
    # Add special tokens for QA
    special_tokens = {
        "additional_special_tokens": ["<question>", "<answer>", "<context>"]
    }
    tokenizer.add_special_tokens(special_tokens)
    model.resize_token_embeddings(len(tokenizer))
    
    # Format QA data
    def format_qa_example(example):
        return f"<context>{example['context']}</context><question>{example['question']}</question><answer>{example['answer']}</answer>"
    
    formatted_data = [format_qa_example(ex) for ex in qa_dataset]
    
    return model, tokenizer, formatted_data
```

#### 2. Text Classification

```python
from transformers import AutoModelForSequenceClassification

def setup_classification_finetuning(model_name, num_labels):
    """Setup fine-tuning for text classification"""
    model = AutoModelForSequenceClassification.from_pretrained(
        model_name,
        num_labels=num_labels
    )
    
    tokenizer = AutoTokenizer.from_pretrained(model_name)
    
    return model, tokenizer
```

### Model Merging and Deployment

#### 1. Merge LoRA Adapters

```python
from peft import PeftModel

def merge_lora_adapters(base_model_name, lora_model_path, output_path):
    """Merge LoRA adapters back into the base model"""
    # Load base model
    base_model = AutoModelForCausalLM.from_pretrained(base_model_name)
    
    # Load LoRA model
    lora_model = PeftModel.from_pretrained(base_model, lora_model_path)
    
    # Merge adapters
    merged_model = lora_model.merge_and_unload()
    
    # Save merged model
    merged_model.save_pretrained(output_path)
    
    return merged_model
```

#### 2. Deploy Fine-tuned Model

```python
import torch
from transformers import pipeline

def deploy_finetuned_model(model_path, task="text-generation"):
    """Deploy fine-tuned model for inference"""
    # Load model and tokenizer
    model = AutoModelForCausalLM.from_pretrained(model_path)
    tokenizer = AutoTokenizer.from_pretrained(model_path)
    
    # Create pipeline
    pipe = pipeline(
        task,
        model=model,
        tokenizer=tokenizer,
        device=0 if torch.cuda.is_available() else -1
    )
    
    return pipe

# Usage
def generate_response(pipe, prompt, max_length=100):
    """Generate response using fine-tuned model"""
    response = pipe(
        prompt,
        max_length=max_length,
        temperature=0.7,
        do_sample=True,
        pad_token_id=pipe.tokenizer.eos_token_id
    )
    return response[0]['generated_text']
```

### Fine-tuning Evaluation

#### 1. Task-Specific Metrics

```python
from sklearn.metrics import accuracy_score, f1_score
import torch

def evaluate_classification(model, test_dataloader, device):
    """Evaluate classification model"""
    model.eval()
    predictions = []
    true_labels = []
    
    with torch.no_grad():
        for batch in test_dataloader:
            batch = {k: v.to(device) for k, v in batch.items()}
            
            outputs = model(**batch)
            logits = outputs.logits
            preds = torch.argmax(logits, dim=-1)
            
            predictions.extend(preds.cpu().numpy())
            true_labels.extend(batch['labels'].cpu().numpy())
    
    accuracy = accuracy_score(true_labels, predictions)
    f1 = f1_score(true_labels, predictions, average='weighted')
    
    return {'accuracy': accuracy, 'f1_score': f1}
```

#### 2. Generation Quality Metrics

```python
from rouge_score import rouge_scorer
import nltk
from nltk.translate.bleu_score import sentence_bleu

def evaluate_generation_quality(generated_texts, reference_texts):
    """Evaluate quality of generated text"""
    rouge = rouge_scorer.RougeScorer(['rouge1', 'rouge2', 'rougeL'], use_stemmer=True)
    
    rouge_scores = []
    bleu_scores = []
    
    for gen, ref in zip(generated_texts, reference_texts):
        # ROUGE scores
        rouge_score = rouge.score(ref, gen)
        rouge_scores.append({
            'rouge1': rouge_score['rouge1'].fmeasure,
            'rouge2': rouge_score['rouge2'].fmeasure,
            'rougeL': rouge_score['rougeL'].fmeasure
        })
        
        # BLEU score
        bleu = sentence_bleu([ref.split()], gen.split())
        bleu_scores.append(bleu)
    
    # Calculate averages
    avg_rouge = {
        metric: np.mean([score[metric] for score in rouge_scores])
        for metric in ['rouge1', 'rouge2', 'rougeL']
    }
    avg_bleu = np.mean(bleu_scores)
    
    return {
        'rouge': avg_rouge,
        'bleu': avg_bleu
    }
```

**In Agent Bruno**:
- **LoRA Fine-tuning**: Uses parameter-efficient methods for domain adaptation
- **Technical Domain**: Fine-tuned on SRE and DevOps documentation
- **Memory Integration**: Learns from user interactions and feedback
- **Continuous Learning**: Updates model based on new conversations
- **Tool Integration**: Fine-tuned to better use MCP tools and protocols

📖 **Related Docs**: [MEMORY.md](./MEMORY.md) | [LEARNING.md](./LEARNING.md) | [OLLAMA.md](./OLLAMA.md)

---

## 7. Vector Databases

### What are Vector Databases?

Vector databases are specialized databases designed to store, index, and search high-dimensional vectors efficiently. They are essential for AI applications that rely on semantic search, similarity matching, and retrieval-augmented generation (RAG).

**Key Features**:
- **High-Dimensional Storage**: Efficiently store vectors with hundreds or thousands of dimensions
- **Similarity Search**: Fast nearest neighbor search using various distance metrics
- **Scalability**: Handle millions of vectors with sub-second query times
- **Metadata Filtering**: Combine vector search with traditional database queries
- **Real-time Updates**: Support for dynamic data insertion and updates

### Vector Database Types

#### 1. Purpose-Built Vector Databases

**Pinecone** (Cloud-native):
```python
import pinecone

# Initialize Pinecone
pinecone.init(api_key="your-api-key", environment="your-environment")

# Create index
pinecone.create_index(
    name="agent-bruno-vectors",
    dimension=384,  # Embedding dimension
    metric="cosine"
)

# Connect to index
index = pinecone.Index("agent-bruno-vectors")

# Insert vectors
vectors_to_insert = [
    ("doc1", [0.1, 0.2, 0.3, ...], {"content": "Kubernetes deployment guide", "type": "document"}),
    ("doc2", [0.4, 0.5, 0.6, ...], {"content": "Docker best practices", "type": "document"})
]
index.upsert(vectors_to_insert)

# Search
query_vector = [0.1, 0.2, 0.3, ...]
results = index.query(
    vector=query_vector,
    top_k=5,
    include_metadata=True,
    filter={"type": "document"}
)
```

**Weaviate** (Open-source):
```python
import weaviate

# Connect to Weaviate
client = weaviate.Client("http://localhost:8080")

# Create schema
schema = {
    "classes": [{
        "class": "Document",
        "description": "Document vectors for Agent Bruno",
        "vectorizer": "none",  # We'll provide vectors
        "properties": [
            {"name": "content", "dataType": ["text"]},
            {"name": "type", "dataType": ["string"]},
            {"name": "source", "dataType": ["string"]}
        ]
    }]
}

client.schema.create(schema)

# Insert data
client.data_object.create(
    data_object={
        "content": "Kubernetes deployment guide",
        "type": "document",
        "source": "k8s-docs"
    },
    class_name="Document",
    vector=[0.1, 0.2, 0.3, ...]  # Your embedding
)

# Search
result = client.query.get("Document", ["content", "type"]).with_near_vector({
    "vector": query_vector
}).with_limit(5).do()
```

#### 2. Traditional Databases with Vector Extensions

**PostgreSQL with pgvector**:
```python
import psycopg2
import numpy as np

# Connect to PostgreSQL
conn = psycopg2.connect(
    host="localhost",
    database="agent_bruno",
    user="postgres",
    password="password"
)

# Create table with vector column
cursor = conn.cursor()
cursor.execute("""
    CREATE TABLE documents (
        id SERIAL PRIMARY KEY,
        content TEXT,
        embedding VECTOR(384),
        metadata JSONB
    );
""")

# Create vector index
cursor.execute("""
    CREATE INDEX ON documents 
    USING ivfflat (embedding vector_cosine_ops) 
    WITH (lists = 100);
""")

# Insert document
embedding = np.array([0.1, 0.2, 0.3, ...])
cursor.execute("""
    INSERT INTO documents (content, embedding, metadata)
    VALUES (%s, %s, %s)
""", ("Kubernetes guide", embedding.tolist(), '{"type": "document"}'))

# Search
cursor.execute("""
    SELECT content, metadata, 
           1 - (embedding <=> %s) as similarity
    FROM documents
    ORDER BY embedding <=> %s
    LIMIT 5
""", (query_embedding.tolist(), query_embedding.tolist()))

results = cursor.fetchall()
```

**Chroma** (Embedded vector database):
```python
import chromadb
from chromadb.config import Settings

# Initialize Chroma
client = chromadb.Client(Settings(
    chroma_db_impl="duckdb+parquet",
    persist_directory="./chroma_db"
))

# Create collection
collection = client.create_collection(
    name="agent_bruno_docs",
    metadata={"description": "Agent Bruno document vectors"}
)

# Add documents
collection.add(
    documents=["Kubernetes deployment guide", "Docker best practices"],
    metadatas=[{"type": "document"}, {"type": "document"}],
    ids=["doc1", "doc2"],
    embeddings=[[0.1, 0.2, 0.3, ...], [0.4, 0.5, 0.6, ...]]
)

# Query
results = collection.query(
    query_texts=["How to deploy Kubernetes?"],
    n_results=5,
    where={"type": "document"}
)
```

### LanceDB Integration (Agent Bruno's Choice)

LanceDB is a serverless vector database that's perfect for Agent Bruno's architecture:

```python
import lancedb
import pyarrow as pa
from typing import List, Dict

class AgentBrunoVectorStore:
    def __init__(self, db_path: str = "./agent_bruno_db"):
        self.db = lancedb.connect(db_path)
        self.table = None
    
    def create_table(self, table_name: str = "documents"):
        """Create a new table for storing document vectors"""
        schema = pa.schema([
            pa.field("id", pa.string()),
            pa.field("content", pa.string()),
            pa.field("embedding", pa.list_(pa.float32())),
            pa.field("metadata", pa.string()),  # JSON string
            pa.field("timestamp", pa.timestamp('us'))
        ])
        
        self.table = self.db.create_table(table_name, schema)
        return self.table
    
    def add_documents(self, documents: List[Dict]):
        """Add documents to the vector store"""
        if not self.table:
            self.create_table()
        
        # Prepare data for insertion
        data = []
        for doc in documents:
            data.append({
                'id': doc['id'],
                'content': doc['content'],
                'embedding': doc['embedding'],
                'metadata': json.dumps(doc.get('metadata', {})),
                'timestamp': pa.scalar(doc.get('timestamp', pa.timestamp('us').value))
            })
        
        self.table.add(data)
    
    def search(self, query_embedding: List[float], top_k: int = 5, 
               filter_metadata: Dict = None) -> List[Dict]:
        """Search for similar documents"""
        if not self.table:
            raise ValueError("Table not created")
        
        # Build query
        query = self.table.search(query_embedding).limit(top_k)
        
        # Add metadata filtering if provided
        if filter_metadata:
            for key, value in filter_metadata.items():
                query = query.where(f"JSON_EXTRACT(metadata, '$.{key}') = '{value}'")
        
        # Execute query
        results = query.to_pandas()
        
        # Convert to list of dictionaries
        documents = []
        for _, row in results.iterrows():
            documents.append({
                'id': row['id'],
                'content': row['content'],
                'metadata': json.loads(row['metadata']),
                'similarity': 1 - row['_distance']  # Convert distance to similarity
            })
        
        return documents
    
    def update_document(self, doc_id: str, content: str = None, 
                       embedding: List[float] = None, metadata: Dict = None):
        """Update an existing document"""
        if not self.table:
            raise ValueError("Table not created")
        
        # Build update data
        update_data = {'id': doc_id}
        if content is not None:
            update_data['content'] = content
        if embedding is not None:
            update_data['embedding'] = embedding
        if metadata is not None:
            update_data['metadata'] = json.dumps(metadata)
        
        # Update document
        self.table.update(update_data, where=f"id = '{doc_id}'")
    
    def delete_document(self, doc_id: str):
        """Delete a document"""
        if not self.table:
            raise ValueError("Table not created")
        
        self.table.delete(where=f"id = '{doc_id}'")
```

### Performance Optimization

#### 1. Indexing Strategies

```python
class OptimizedVectorStore:
    def __init__(self, db_path: str):
        self.db = lancedb.connect(db_path)
        self.table = None
    
    def create_optimized_table(self, table_name: str):
        """Create table with optimized settings"""
        schema = pa.schema([
            pa.field("id", pa.string()),
            pa.field("content", pa.string()),
            pa.field("embedding", pa.list_(pa.float32())),
            pa.field("metadata", pa.string()),
            pa.field("timestamp", pa.timestamp('us')),
            pa.field("content_hash", pa.string())  # For deduplication
        ])
        
        self.table = self.db.create_table(
            table_name, 
            schema,
            # Optimize for search performance
            index_type="IVF_PQ",  # Product quantization
            num_partitions=100,
            num_sub_vectors=16
        )
        
        return self.table
    
    def batch_insert(self, documents: List[Dict], batch_size: int = 1000):
        """Insert documents in batches for better performance"""
        for i in range(0, len(documents), batch_size):
            batch = documents[i:i + batch_size]
            self.table.add(batch)
    
    def create_metadata_index(self, metadata_field: str):
        """Create index on metadata field for faster filtering"""
        # This would depend on the specific vector database implementation
        pass
```

#### 2. Caching and Preprocessing

```python
import hashlib
import pickle
from typing import Optional

class CachedVectorStore:
    def __init__(self, vector_store, cache_size: int = 10000):
        self.vector_store = vector_store
        self.cache = {}
        self.cache_size = cache_size
    
    def search_with_cache(self, query_embedding: List[float], 
                         top_k: int = 5, cache_key: str = None) -> List[Dict]:
        """Search with caching for repeated queries"""
        # Generate cache key
        if cache_key is None:
            cache_key = hashlib.md5(str(query_embedding).encode()).hexdigest()
        
        # Check cache
        if cache_key in self.cache:
            return self.cache[cache_key]
        
        # Perform search
        results = self.vector_store.search(query_embedding, top_k)
        
        # Cache results
        if len(self.cache) >= self.cache_size:
            # Remove oldest entry
            oldest_key = next(iter(self.cache))
            del self.cache[oldest_key]
        
        self.cache[cache_key] = results
        return results
    
    def precompute_embeddings(self, texts: List[str], 
                             embedding_model) -> List[List[float]]:
        """Precompute embeddings for better performance"""
        embeddings = []
        for text in texts:
            embedding = embedding_model.encode([text])[0]
            embeddings.append(embedding.tolist())
        return embeddings
```

### Advanced Vector Database Features

#### 1. Hybrid Search

```python
class HybridVectorStore:
    def __init__(self, vector_store, text_index):
        self.vector_store = vector_store
        self.text_index = text_index  # Elasticsearch or similar
    
    def hybrid_search(self, query: str, query_embedding: List[float], 
                     top_k: int = 10, alpha: float = 0.7) -> List[Dict]:
        """Combine vector search with text search"""
        # Vector search
        vector_results = self.vector_store.search(query_embedding, top_k * 2)
        
        # Text search
        text_results = self.text_index.search(query, top_k * 2)
        
        # Combine results
        combined_results = self._combine_results(
            vector_results, text_results, alpha
        )
        
        return combined_results[:top_k]
    
    def _combine_results(self, vector_results: List[Dict], 
                        text_results: List[Dict], alpha: float) -> List[Dict]:
        """Combine vector and text search results"""
        # Implementation would depend on specific scoring methods
        pass
```

#### 2. Real-time Updates

```python
class RealTimeVectorStore:
    def __init__(self, vector_store, event_stream):
        self.vector_store = vector_store
        self.event_stream = event_stream
    
    def setup_real_time_updates(self):
        """Setup real-time updates from event stream"""
        for event in self.event_stream:
            if event['type'] == 'document_added':
                self.vector_store.add_documents([event['document']])
            elif event['type'] == 'document_updated':
                self.vector_store.update_document(
                    event['doc_id'], 
                    event['content'],
                    event['embedding']
                )
            elif event['type'] == 'document_deleted':
                self.vector_store.delete_document(event['doc_id'])
```

### Monitoring and Maintenance

#### 1. Performance Monitoring

```python
import time
from typing import Dict, List

class MonitoredVectorStore:
    def __init__(self, vector_store):
        self.vector_store = vector_store
        self.metrics = {
            'search_times': [],
            'insert_times': [],
            'query_counts': 0,
            'error_counts': 0
        }
    
    def search(self, query_embedding: List[float], top_k: int = 5) -> List[Dict]:
        """Search with performance monitoring"""
        start_time = time.time()
        
        try:
            results = self.vector_store.search(query_embedding, top_k)
            self.metrics['query_counts'] += 1
        except Exception as e:
            self.metrics['error_counts'] += 1
            raise e
        finally:
            search_time = time.time() - start_time
            self.metrics['search_times'].append(search_time)
        
        return results
    
    def get_performance_metrics(self) -> Dict:
        """Get performance metrics"""
        if not self.metrics['search_times']:
            return self.metrics
        
        return {
            'avg_search_time': sum(self.metrics['search_times']) / len(self.metrics['search_times']),
            'max_search_time': max(self.metrics['search_times']),
            'min_search_time': min(self.metrics['search_times']),
            'query_counts': self.metrics['query_counts'],
            'error_counts': self.metrics['error_counts'],
            'error_rate': self.metrics['error_counts'] / max(self.metrics['query_counts'], 1)
        }
```

#### 2. Data Quality Monitoring

```python
class QualityMonitoredVectorStore:
    def __init__(self, vector_store):
        self.vector_store = vector_store
        self.quality_metrics = {
            'duplicate_documents': 0,
            'empty_embeddings': 0,
            'dimension_mismatches': 0
        }
    
    def add_documents_with_quality_check(self, documents: List[Dict]):
        """Add documents with quality checks"""
        validated_documents = []
        
        for doc in documents:
            # Check for duplicates
            if self._is_duplicate(doc):
                self.quality_metrics['duplicate_documents'] += 1
                continue
            
            # Check embedding quality
            if not self._is_valid_embedding(doc.get('embedding')):
                self.quality_metrics['empty_embeddings'] += 1
                continue
            
            # Check dimension consistency
            if not self._has_consistent_dimensions(doc.get('embedding')):
                self.quality_metrics['dimension_mismatches'] += 1
                continue
            
            validated_documents.append(doc)
        
        if validated_documents:
            self.vector_store.add_documents(validated_documents)
    
    def _is_duplicate(self, doc: Dict) -> bool:
        """Check if document is duplicate"""
        # Implementation would check against existing documents
        return False
    
    def _is_valid_embedding(self, embedding: List[float]) -> bool:
        """Check if embedding is valid"""
        return embedding is not None and len(embedding) > 0
    
    def _has_consistent_dimensions(self, embedding: List[float]) -> bool:
        """Check if embedding has consistent dimensions"""
        expected_dim = 384  # Your embedding dimension
        return len(embedding) == expected_dim
```

**In Agent Bruno**:
- **LanceDB Integration**: Uses LanceDB for embedded vector storage
- **Hybrid Search**: Combines semantic and keyword search
- **Real-time Updates**: Updates vectors based on new conversations
- **Memory Persistence**: Stores conversation embeddings for long-term memory
- **Performance Optimization**: Optimized for fast similarity search
- **Metadata Filtering**: Supports filtering by document type, source, etc.

📖 **Related Docs**: [LANCEDB_PERSISTENCE.md](./LANCEDB_PERSISTENCE.md) | [RAG.md](./RAG.md) | [MEMORY.md](./MEMORY.md)

---

## 8. Production ML Engineering

### What is MLOps?

MLOps (Machine Learning Operations) is the practice of applying DevOps principles to machine learning systems. It encompasses the entire ML lifecycle from development to deployment, monitoring, and maintenance.

**Key MLOps Principles**:
- **Reproducibility**: Ensure experiments can be reproduced
- **Versioning**: Track data, code, and model versions
- **Automation**: Automate training, testing, and deployment
- **Monitoring**: Track model performance in production
- **Governance**: Ensure compliance and security
- **Collaboration**: Enable team collaboration on ML projects

### ML Pipeline Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    ML Pipeline Architecture                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Data Sources → Data Validation → Feature Engineering          │
│       ↓              ↓                    ↓                     │
│  Data Storage → Feature Store → Model Training                 │
│       ↓              ↓                    ↓                     │
│  Data Lineage → Model Registry → Model Validation              │
│       ↓              ↓                    ↓                     │
│  Monitoring → Model Serving → A/B Testing                     │
│       ↓              ↓                    ↓                     │
│  Feedback Loop → Model Retraining → Model Deployment          │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Data Management

#### 1. Data Versioning

```python
import dvc.api
import pandas as pd
from typing import Dict, Any

class DataVersioning:
    def __init__(self, repo_path: str):
        self.repo_path = repo_path
    
    def track_data(self, data_path: str, metadata: Dict[str, Any]):
        """Track data with DVC"""
        # Add data to DVC
        dvc.api.add(data_path)
        
        # Commit with metadata
        import subprocess
        subprocess.run(['git', 'add', f'{data_path}.dvc'])
        subprocess.run(['git', 'commit', '-m', f'Add {data_path} with metadata: {metadata}'])
    
    def get_data_version(self, data_path: str, version: str = None):
        """Get specific version of data"""
        if version:
            return dvc.api.get_url(data_path, repo=self.repo_path, rev=version)
        else:
            return dvc.api.get_url(data_path, repo=self.repo_path)
    
    def list_data_versions(self, data_path: str):
        """List all versions of a data file"""
        # Implementation would use git log to find commits
        pass
```

#### 2. Data Validation

```python
from great_expectations import DataContext
import pandas as pd

class DataValidator:
    def __init__(self, context_path: str):
        self.context = DataContext(context_path)
    
    def validate_data(self, df: pd.DataFrame, expectation_suite: str):
        """Validate data against expectations"""
        validator = self.context.get_validator(
            batch_request={
                "datasource_name": "my_datasource",
                "data_connector_name": "default_inferred_data_connector_name",
                "data_asset_name": "my_data_asset",
            },
            expectation_suite_name=expectation_suite,
        )
        
        # Run validation
        validation_result = validator.validate()
        
        return validation_result
    
    def create_expectation_suite(self, df: pd.DataFrame, suite_name: str):
        """Create expectation suite from data sample"""
        validator = self.context.get_validator(
            batch_request={
                "datasource_name": "my_datasource",
                "data_connector_name": "default_inferred_data_connector_name",
                "data_asset_name": "my_data_asset",
            }
        )
        
        # Add expectations
        validator.expect_column_to_exist("target")
        validator.expect_column_values_to_not_be_null("target")
        validator.expect_column_values_to_be_between("target", min_value=0, max_value=1)
        
        # Save suite
        validator.save_expectation_suite(discard_failed_expectations=False)
```

### Feature Engineering

#### 1. Feature Store

```python
from feast import FeatureStore
import pandas as pd
from datetime import datetime

class FeatureStoreManager:
    def __init__(self, repo_path: str):
        self.store = FeatureStore(repo_path=repo_path)
    
    def create_feature_view(self, name: str, features: list, entity: str):
        """Create a feature view"""
        from feast import FeatureView, Field
        from feast.types import Float64, String
        
        feature_view = FeatureView(
            name=name,
            entities=[entity],
            schema=[
                Field(name=feature, dtype=Float64) for feature in features
            ],
            source=self.store.get_data_source("my_source"),
            ttl=timedelta(days=1)
        )
        
        return feature_view
    
    def get_historical_features(self, entity_df: pd.DataFrame, features: list):
        """Get historical features for training"""
        training_df = self.store.get_historical_features(
            entity_df=entity_df,
            features=features
        ).to_df()
        
        return training_df
    
    def get_online_features(self, entity_rows: list, features: list):
        """Get online features for inference"""
        online_features = self.store.get_online_features(
            entity_rows=entity_rows,
            features=features
        ).to_dict()
        
        return online_features
```

#### 2. Feature Engineering Pipeline

```python
from sklearn.preprocessing import StandardScaler, LabelEncoder
from sklearn.feature_selection import SelectKBest, f_classif
import pandas as pd

class FeatureEngineeringPipeline:
    def __init__(self):
        self.scaler = StandardScaler()
        self.encoder = LabelEncoder()
        self.feature_selector = SelectKBest(f_classif, k=10)
        self.is_fitted = False
    
    def fit_transform(self, X: pd.DataFrame, y: pd.Series = None):
        """Fit and transform features"""
        # Handle categorical variables
        categorical_cols = X.select_dtypes(include=['object']).columns
        for col in categorical_cols:
            X[col] = self.encoder.fit_transform(X[col])
        
        # Scale numerical features
        X_scaled = self.scaler.fit_transform(X)
        
        # Select best features if target is provided
        if y is not None:
            X_selected = self.feature_selector.fit_transform(X_scaled, y)
            self.is_fitted = True
            return X_selected
        
        self.is_fitted = True
        return X_scaled
    
    def transform(self, X: pd.DataFrame):
        """Transform new data using fitted pipeline"""
        if not self.is_fitted:
            raise ValueError("Pipeline must be fitted first")
        
        # Handle categorical variables
        categorical_cols = X.select_dtypes(include=['object']).columns
        for col in categorical_cols:
            X[col] = self.encoder.transform(X[col])
        
        # Scale features
        X_scaled = self.scaler.transform(X)
        
        # Select features
        X_selected = self.feature_selector.transform(X_scaled)
        
        return X_selected
```

### Model Management

#### 1. Model Registry

```python
import mlflow
import mlflow.sklearn
from mlflow.tracking import MlflowClient
import joblib

class ModelRegistry:
    def __init__(self, tracking_uri: str):
        mlflow.set_tracking_uri(tracking_uri)
        self.client = MlflowClient()
    
    def log_model(self, model, model_name: str, metrics: dict, 
                  params: dict, tags: dict = None):
        """Log model to MLflow"""
        with mlflow.start_run():
            # Log parameters
            for key, value in params.items():
                mlflow.log_param(key, value)
            
            # Log metrics
            for key, value in metrics.items():
                mlflow.log_metric(key, value)
            
            # Log model
            mlflow.sklearn.log_model(
                model, 
                "model",
                registered_model_name=model_name
            )
            
            # Log tags
            if tags:
                for key, value in tags.items():
                    mlflow.set_tag(key, value)
    
    def get_model_versions(self, model_name: str):
        """Get all versions of a model"""
        return self.client.search_model_versions(f"name='{model_name}'")
    
    def promote_model(self, model_name: str, version: str, stage: str):
        """Promote model to a specific stage"""
        self.client.transition_model_version_stage(
            name=model_name,
            version=version,
            stage=stage
        )
    
    def load_model(self, model_name: str, stage: str = "Production"):
        """Load model from registry"""
        model_uri = f"models:/{model_name}/{stage}"
        return mlflow.sklearn.load_model(model_uri)
```

#### 2. Model Validation

```python
from sklearn.metrics import accuracy_score, precision_score, recall_score, f1_score
import numpy as np

class ModelValidator:
    def __init__(self, validation_thresholds: dict):
        self.thresholds = validation_thresholds
    
    def validate_model(self, model, X_test: np.ndarray, y_test: np.ndarray) -> dict:
        """Validate model performance"""
        # Make predictions
        y_pred = model.predict(X_test)
        y_pred_proba = model.predict_proba(X_test)[:, 1] if hasattr(model, 'predict_proba') else None
        
        # Calculate metrics
        metrics = {
            'accuracy': accuracy_score(y_test, y_pred),
            'precision': precision_score(y_test, y_pred, average='weighted'),
            'recall': recall_score(y_test, y_pred, average='weighted'),
            'f1_score': f1_score(y_test, y_pred, average='weighted')
        }
        
        # Check if model meets validation criteria
        validation_passed = all(
            metrics[metric] >= threshold 
            for metric, threshold in self.thresholds.items()
        )
        
        return {
            'metrics': metrics,
            'validation_passed': validation_passed,
            'thresholds': self.thresholds
        }
    
    def validate_data_drift(self, X_train: np.ndarray, X_test: np.ndarray) -> dict:
        """Validate for data drift"""
        from scipy import stats
        
        drift_results = {}
        
        for i in range(X_train.shape[1]):
            # Kolmogorov-Smirnov test
            ks_stat, ks_pvalue = stats.ks_2samp(X_train[:, i], X_test[:, i])
            
            drift_results[f'feature_{i}'] = {
                'ks_statistic': ks_stat,
                'ks_pvalue': ks_pvalue,
                'drift_detected': ks_pvalue < 0.05
            }
        
        return drift_results
```

### Model Deployment

#### 1. Model Serving

```python
from fastapi import FastAPI, HTTPException
import joblib
import numpy as np
from pydantic import BaseModel
import logging

class PredictionRequest(BaseModel):
    features: list
    model_version: str = "latest"

class ModelServer:
    def __init__(self, model_path: str):
        self.app = FastAPI(title="ML Model Server")
        self.model = joblib.load(model_path)
        self.logger = logging.getLogger(__name__)
        
        # Setup routes
        self.setup_routes()
    
    def setup_routes(self):
        @self.app.post("/predict")
        async def predict(request: PredictionRequest):
            try:
                # Validate input
                if len(request.features) != self.model.n_features_in_:
                    raise HTTPException(
                        status_code=400,
                        detail=f"Expected {self.model.n_features_in_} features, got {len(request.features)}"
                    )
                
                # Make prediction
                prediction = self.model.predict([request.features])[0]
                probability = self.model.predict_proba([request.features])[0] if hasattr(self.model, 'predict_proba') else None
                
                return {
                    "prediction": prediction,
                    "probability": probability.tolist() if probability is not None else None,
                    "model_version": request.model_version
                }
                
            except Exception as e:
                self.logger.error(f"Prediction error: {str(e)}")
                raise HTTPException(status_code=500, detail=str(e))
        
        @self.app.get("/health")
        async def health_check():
            return {"status": "healthy"}
    
    def run(self, host: str = "0.0.0.0", port: int = 8000):
        import uvicorn
        uvicorn.run(self.app, host=host, port=port)
```

#### 2. A/B Testing

```python
import random
from typing import Dict, Any
import logging

class ABTestingFramework:
    def __init__(self, models: Dict[str, Any], traffic_split: Dict[str, float]):
        self.models = models
        self.traffic_split = traffic_split
        self.logger = logging.getLogger(__name__)
        
        # Validate traffic split
        if abs(sum(traffic_split.values()) - 1.0) > 1e-6:
            raise ValueError("Traffic split must sum to 1.0")
    
    def get_model_for_request(self, user_id: str) -> str:
        """Get model assignment for a user"""
        # Use consistent hashing for stable assignments
        hash_value = hash(user_id) % 100
        
        cumulative = 0
        for model_name, split in self.traffic_split.items():
            cumulative += split * 100
            if hash_value < cumulative:
                return model_name
        
        # Fallback to first model
        return list(self.traffic_split.keys())[0]
    
    def predict(self, user_id: str, features: list) -> Dict[str, Any]:
        """Make prediction using assigned model"""
        model_name = self.get_model_for_request(user_id)
        model = self.models[model_name]
        
        # Make prediction
        prediction = model.predict([features])[0]
        
        # Log for analysis
        self.logger.info(f"User {user_id} assigned to {model_name}, prediction: {prediction}")
        
        return {
            "prediction": prediction,
            "model_used": model_name,
            "user_id": user_id
        }
    
    def update_traffic_split(self, new_split: Dict[str, float]):
        """Update traffic split for models"""
        if abs(sum(new_split.values()) - 1.0) > 1e-6:
            raise ValueError("Traffic split must sum to 1.0")
        
        self.traffic_split = new_split
        self.logger.info(f"Updated traffic split: {new_split}")
```

### Monitoring and Observability

#### 1. Model Performance Monitoring

```python
import time
from typing import Dict, List
import numpy as np
from collections import defaultdict, deque

class ModelMonitor:
    def __init__(self, window_size: int = 1000):
        self.window_size = window_size
        self.predictions = deque(maxlen=window_size)
        self.latencies = deque(maxlen=window_size)
        self.errors = deque(maxlen=window_size)
        self.performance_metrics = defaultdict(list)
    
    def log_prediction(self, prediction: Any, latency: float, error: bool = False):
        """Log a prediction for monitoring"""
        self.predictions.append(prediction)
        self.latencies.append(latency)
        self.errors.append(error)
    
    def get_performance_metrics(self) -> Dict[str, float]:
        """Get current performance metrics"""
        if not self.predictions:
            return {}
        
        return {
            'avg_latency': np.mean(self.latencies),
            'p95_latency': np.percentile(self.latencies, 95),
            'p99_latency': np.percentile(self.latencies, 99),
            'error_rate': np.mean(self.errors),
            'throughput': len(self.predictions) / (time.time() - self.start_time) if hasattr(self, 'start_time') else 0
        }
    
    def detect_drift(self, reference_data: np.ndarray, threshold: float = 0.05) -> bool:
        """Detect data drift using statistical tests"""
        if len(self.predictions) < 100:
            return False
        
        current_data = np.array(self.predictions)
        
        # Kolmogorov-Smirnov test
        from scipy import stats
        ks_stat, p_value = stats.ks_2samp(reference_data, current_data)
        
        return p_value < threshold
    
    def alert_on_anomalies(self, thresholds: Dict[str, float]) -> List[str]:
        """Check for performance anomalies"""
        alerts = []
        metrics = self.get_performance_metrics()
        
        for metric, threshold in thresholds.items():
            if metric in metrics and metrics[metric] > threshold:
                alerts.append(f"{metric} exceeded threshold: {metrics[metric]:.3f} > {threshold}")
        
        return alerts
```

#### 2. Data Quality Monitoring

```python
import pandas as pd
from typing import Dict, List, Any

class DataQualityMonitor:
    def __init__(self, reference_data: pd.DataFrame):
        self.reference_data = reference_data
        self.quality_metrics = defaultdict(list)
    
    def check_data_quality(self, new_data: pd.DataFrame) -> Dict[str, Any]:
        """Check data quality against reference"""
        quality_report = {}
        
        # Check for missing values
        missing_values = new_data.isnull().sum()
        quality_report['missing_values'] = missing_values.to_dict()
        
        # Check for data types
        type_changes = {}
        for col in new_data.columns:
            if col in self.reference_data.columns:
                if new_data[col].dtype != self.reference_data[col].dtype:
                    type_changes[col] = {
                        'expected': str(self.reference_data[col].dtype),
                        'actual': str(new_data[col].dtype)
                    }
        quality_report['type_changes'] = type_changes
        
        # Check for outliers
        outliers = {}
        for col in new_data.select_dtypes(include=[np.number]).columns:
            if col in self.reference_data.columns:
                Q1 = self.reference_data[col].quantile(0.25)
                Q3 = self.reference_data[col].quantile(0.75)
                IQR = Q3 - Q1
                lower_bound = Q1 - 1.5 * IQR
                upper_bound = Q3 + 1.5 * IQR
                
                outlier_count = ((new_data[col] < lower_bound) | (new_data[col] > upper_bound)).sum()
                outliers[col] = outlier_count
        quality_report['outliers'] = outliers
        
        # Check for distribution changes
        distribution_changes = {}
        for col in new_data.select_dtypes(include=[np.number]).columns:
            if col in self.reference_data.columns:
                from scipy import stats
                ks_stat, p_value = stats.ks_2samp(
                    self.reference_data[col].dropna(),
                    new_data[col].dropna()
                )
                distribution_changes[col] = {
                    'ks_statistic': ks_stat,
                    'p_value': p_value,
                    'significant_change': p_value < 0.05
                }
        quality_report['distribution_changes'] = distribution_changes
        
        return quality_report
```

### CI/CD for ML

#### 1. ML Pipeline Automation

```python
import subprocess
import yaml
from typing import Dict, Any

class MLPipeline:
    def __init__(self, config_path: str):
        with open(config_path, 'r') as f:
            self.config = yaml.safe_load(f)
    
    def run_data_pipeline(self):
        """Run data processing pipeline"""
        steps = self.config['data_pipeline']['steps']
        
        for step in steps:
            print(f"Running step: {step['name']}")
            
            # Run data validation
            if step['type'] == 'validation':
                self.run_data_validation(step['config'])
            
            # Run feature engineering
            elif step['type'] == 'feature_engineering':
                self.run_feature_engineering(step['config'])
            
            # Run data quality checks
            elif step['type'] == 'quality_check':
                self.run_quality_check(step['config'])
    
    def run_training_pipeline(self):
        """Run model training pipeline"""
        training_config = self.config['training']
        
        # Train model
        print("Training model...")
        subprocess.run(['python', 'train_model.py', '--config', training_config['config_path']])
        
        # Validate model
        print("Validating model...")
        subprocess.run(['python', 'validate_model.py', '--config', training_config['validation_config']])
        
        # Register model
        print("Registering model...")
        subprocess.run(['python', 'register_model.py', '--config', training_config['registry_config']])
    
    def run_deployment_pipeline(self):
        """Run model deployment pipeline"""
        deployment_config = self.config['deployment']
        
        # Build Docker image
        print("Building Docker image...")
        subprocess.run(['docker', 'build', '-t', deployment_config['image_name'], '.'])
        
        # Run tests
        print("Running tests...")
        subprocess.run(['python', '-m', 'pytest', 'tests/'])
        
        # Deploy to staging
        print("Deploying to staging...")
        subprocess.run(['kubectl', 'apply', '-f', 'k8s/staging/'])
        
        # Run integration tests
        print("Running integration tests...")
        subprocess.run(['python', 'tests/integration_tests.py'])
        
        # Deploy to production
        print("Deploying to production...")
        subprocess.run(['kubectl', 'apply', '-f', 'k8s/production/'])
```

#### 2. Automated Testing

```python
import pytest
import numpy as np
from sklearn.model_selection import train_test_split
from sklearn.ensemble import RandomForestClassifier

class TestMLPipeline:
    def test_data_quality(self, sample_data):
        """Test data quality"""
        # Check for missing values
        assert sample_data.isnull().sum().sum() == 0, "Data contains missing values"
        
        # Check for duplicates
        assert sample_data.duplicated().sum() == 0, "Data contains duplicates"
        
        # Check data types
        expected_types = {'feature1': 'float64', 'feature2': 'int64', 'target': 'int64'}
        for col, expected_type in expected_types.items():
            assert sample_data[col].dtype == expected_type, f"Column {col} has wrong type"
    
    def test_model_performance(self, trained_model, test_data, test_labels):
        """Test model performance"""
        predictions = trained_model.predict(test_data)
        accuracy = (predictions == test_labels).mean()
        
        assert accuracy >= 0.8, f"Model accuracy {accuracy} is below threshold"
    
    def test_model_bias(self, trained_model, test_data, test_labels, sensitive_features):
        """Test for model bias"""
        predictions = trained_model.predict(test_data)
        
        # Check for bias across sensitive features
        for feature in sensitive_features:
            unique_values = test_data[feature].unique()
            accuracies = []
            
            for value in unique_values:
                mask = test_data[feature] == value
                if mask.sum() > 0:
                    accuracy = (predictions[mask] == test_labels[mask]).mean()
                    accuracies.append(accuracy)
            
            # Check if accuracy varies significantly across groups
            if len(accuracies) > 1:
                max_diff = max(accuracies) - min(accuracies)
                assert max_diff < 0.1, f"Model shows bias across {feature}"
    
    def test_model_robustness(self, trained_model, test_data):
        """Test model robustness"""
        # Test with slightly perturbed data
        perturbed_data = test_data + np.random.normal(0, 0.01, test_data.shape)
        
        original_predictions = trained_model.predict(test_data)
        perturbed_predictions = trained_model.predict(perturbed_data)
        
        # Check if predictions are stable
        prediction_diff = np.abs(original_predictions - perturbed_predictions).mean()
        assert prediction_diff < 0.1, "Model is not robust to small perturbations"
```

**In Agent Bruno**:
- **MLOps Integration**: Uses MLflow for model versioning and tracking
- **Automated Pipelines**: CI/CD for model training and deployment
- **Performance Monitoring**: Tracks model performance and data drift
- **A/B Testing**: Tests different model versions with users
- **Feature Store**: Manages features for training and inference
- **Model Registry**: Tracks model versions and metadata

📖 **Related Docs**: [OBSERVABILITY.md](./OBSERVABILITY.md) | [PRODUCTION_READINESS.md](./PRODUCTION_READINESS.md) | [TESTING.md](./TESTING.md)

---

## 9. AI Observability

### What is AI Observability?

AI Observability is the practice of monitoring, debugging, and understanding AI systems in production. It goes beyond traditional monitoring to provide insights into model behavior, data quality, and system performance.

**Key Components**:
- **Model Performance Monitoring**: Track accuracy, latency, and throughput
- **Data Drift Detection**: Identify when input data changes significantly
- **Model Drift Detection**: Detect when model performance degrades
- **Explainability**: Understand why models make specific predictions
- **Debugging**: Identify and fix issues in AI systems
- **Alerting**: Proactive notification of problems

### Observability Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    AI Observability Stack                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Data Sources → Metrics Collection → Analysis & Alerting       │
│       ↓              ↓                    ↓                     │
│  Model Logs → Time Series DB → Dashboards & Reports            │
│       ↓              ↓                    ↓                     │
│  User Feedback → Feature Store → Model Retraining              │
│       ↓              ↓                    ↓                     │
│  System Metrics → Log Aggregation → Incident Response          │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Model Performance Monitoring

#### 1. Real-time Metrics Collection

```python
import time
import logging
from typing import Dict, Any, List
from collections import defaultdict, deque
import numpy as np

class ModelPerformanceMonitor:
    def __init__(self, window_size: int = 1000):
        self.window_size = window_size
        self.metrics = {
            'predictions': deque(maxlen=window_size),
            'latencies': deque(maxlen=window_size),
            'errors': deque(maxlen=window_size),
            'confidence_scores': deque(maxlen=window_size),
            'input_features': deque(maxlen=window_size)
        }
        self.logger = logging.getLogger(__name__)
    
    def log_prediction(self, prediction: Any, latency: float, 
                      confidence: float = None, input_features: Dict = None,
                      error: bool = False):
        """Log a prediction with associated metrics"""
        self.metrics['predictions'].append(prediction)
        self.metrics['latencies'].append(latency)
        self.metrics['errors'].append(error)
        
        if confidence is not None:
            self.metrics['confidence_scores'].append(confidence)
        
        if input_features is not None:
            self.metrics['input_features'].append(input_features)
    
    def get_performance_summary(self) -> Dict[str, Any]:
        """Get comprehensive performance summary"""
        if not self.metrics['predictions']:
            return {}
        
        return {
            'total_predictions': len(self.metrics['predictions']),
            'avg_latency': np.mean(self.metrics['latencies']),
            'p95_latency': np.percentile(self.metrics['latencies'], 95),
            'p99_latency': np.percentile(self.metrics['latencies'], 99),
            'error_rate': np.mean(self.metrics['errors']),
            'avg_confidence': np.mean(self.metrics['confidence_scores']) if self.metrics['confidence_scores'] else None,
            'throughput': len(self.metrics['predictions']) / (time.time() - self.start_time) if hasattr(self, 'start_time') else 0
        }
    
    def detect_performance_degradation(self, baseline_metrics: Dict[str, float], 
                                     threshold: float = 0.1) -> List[str]:
        """Detect performance degradation compared to baseline"""
        current_metrics = self.get_performance_summary()
        alerts = []
        
        for metric, baseline_value in baseline_metrics.items():
            if metric in current_metrics:
                current_value = current_metrics[metric]
                degradation = abs(current_value - baseline_value) / baseline_value
                
                if degradation > threshold:
                    alerts.append(f"{metric} degraded by {degradation:.2%}")
        
        return alerts
```

#### 2. Model Drift Detection

```python
from scipy import stats
import numpy as np
from typing import Dict, Any

class ModelDriftDetector:
    def __init__(self, reference_data: np.ndarray, reference_predictions: np.ndarray):
        self.reference_data = reference_data
        self.reference_predictions = reference_predictions
        self.drift_threshold = 0.05
    
    def detect_data_drift(self, current_data: np.ndarray) -> Dict[str, Any]:
        """Detect data drift using statistical tests"""
        drift_results = {}
        
        for i in range(current_data.shape[1]):
            # Kolmogorov-Smirnov test
            ks_stat, ks_pvalue = stats.ks_2samp(
                self.reference_data[:, i], 
                current_data[:, i]
            )
            
            # Population Stability Index (PSI)
            psi = self.calculate_psi(
                self.reference_data[:, i], 
                current_data[:, i]
            )
            
            drift_results[f'feature_{i}'] = {
                'ks_statistic': ks_stat,
                'ks_pvalue': ks_pvalue,
                'psi': psi,
                'drift_detected': ks_pvalue < self.drift_threshold or psi > 0.2
            }
        
        return drift_results
    
    def detect_prediction_drift(self, current_predictions: np.ndarray) -> Dict[str, Any]:
        """Detect prediction drift"""
        # Statistical test for prediction distribution
        ks_stat, ks_pvalue = stats.ks_2samp(
            self.reference_predictions, 
            current_predictions
        )
        
        # Calculate PSI for predictions
        psi = self.calculate_psi(
            self.reference_predictions, 
            current_predictions
        )
        
        return {
            'ks_statistic': ks_stat,
            'ks_pvalue': ks_pvalue,
            'psi': psi,
            'drift_detected': ks_pvalue < self.drift_threshold or psi > 0.2
        }
    
    def calculate_psi(self, reference: np.ndarray, current: np.ndarray, 
                      bins: int = 10) -> float:
        """Calculate Population Stability Index"""
        # Create bins based on reference data
        bin_edges = np.linspace(reference.min(), reference.max(), bins + 1)
        
        # Calculate distributions
        ref_hist, _ = np.histogram(reference, bins=bin_edges)
        curr_hist, _ = np.histogram(current, bins=bin_edges)
        
        # Normalize to probabilities
        ref_prob = ref_hist / ref_hist.sum()
        curr_prob = curr_hist / curr_hist.sum()
        
        # Calculate PSI
        psi = np.sum((curr_prob - ref_prob) * np.log(curr_prob / ref_prob))
        
        return psi
```

### Explainability and Interpretability

#### 1. SHAP Integration

```python
import shap
import numpy as np
from typing import Dict, Any, List

class ModelExplainer:
    def __init__(self, model, background_data: np.ndarray):
        self.model = model
        self.background_data = background_data
        self.explainer = shap.Explainer(model, background_data)
    
    def explain_prediction(self, input_data: np.ndarray) -> Dict[str, Any]:
        """Generate explanation for a single prediction"""
        shap_values = self.explainer(input_data)
        
        # Get feature importance
        feature_importance = np.abs(shap_values.values).mean(axis=0)
        
        # Get top contributing features
        top_features = np.argsort(feature_importance)[::-1][:5]
        
        return {
            'shap_values': shap_values.values.tolist(),
            'feature_importance': feature_importance.tolist(),
            'top_features': top_features.tolist(),
            'prediction': self.model.predict(input_data)[0]
        }
    
    def explain_batch(self, input_data: np.ndarray) -> Dict[str, Any]:
        """Generate explanations for a batch of predictions"""
        shap_values = self.explainer(input_data)
        
        # Aggregate explanations
        mean_shap = np.mean(shap_values.values, axis=0)
        std_shap = np.std(shap_values.values, axis=0)
        
        return {
            'mean_shap_values': mean_shap.tolist(),
            'std_shap_values': std_shap.tolist(),
            'feature_importance': np.abs(mean_shap).tolist()
        }
    
    def generate_global_explanation(self, input_data: np.ndarray) -> Dict[str, Any]:
        """Generate global explanation for the model"""
        shap_values = self.explainer(input_data)
        
        # Global feature importance
        global_importance = np.abs(shap_values.values).mean(axis=0)
        
        # Feature interactions
        interactions = shap_values.interaction_values if hasattr(shap_values, 'interaction_values') else None
        
        return {
            'global_importance': global_importance.tolist(),
            'interactions': interactions.tolist() if interactions is not None else None
        }
```

#### 2. LIME Integration

```python
from lime import lime_tabular
import numpy as np

class LIMEModelExplainer:
    def __init__(self, model, training_data: np.ndarray, feature_names: List[str]):
        self.model = model
        self.feature_names = feature_names
        self.explainer = lime_tabular.LimeTabularExplainer(
            training_data,
            feature_names=feature_names,
            mode='classification'
        )
    
    def explain_prediction(self, input_data: np.ndarray, num_features: int = 5) -> Dict[str, Any]:
        """Generate LIME explanation for a single prediction"""
        explanation = self.explainer.explain_instance(
            input_data[0], 
            self.model.predict_proba,
            num_features=num_features
        )
        
        # Extract explanation details
        explanation_dict = {
            'prediction': self.model.predict(input_data)[0],
            'explanation': explanation.as_list(),
            'local_prediction': explanation.local_pred,
            'score': explanation.score
        }
        
        return explanation_dict
```

### Data Quality Monitoring

#### 1. Data Quality Metrics

```python
import pandas as pd
from typing import Dict, Any, List
import numpy as np

class DataQualityMonitor:
    def __init__(self, reference_data: pd.DataFrame):
        self.reference_data = reference_data
        self.quality_thresholds = {
            'missing_rate': 0.05,
            'outlier_rate': 0.1,
            'distribution_shift': 0.05
        }
    
    def check_data_quality(self, current_data: pd.DataFrame) -> Dict[str, Any]:
        """Comprehensive data quality check"""
        quality_report = {
            'timestamp': pd.Timestamp.now(),
            'data_shape': current_data.shape,
            'quality_issues': []
        }
        
        # Check for missing values
        missing_report = self.check_missing_values(current_data)
        quality_report['missing_values'] = missing_report
        
        # Check for outliers
        outlier_report = self.check_outliers(current_data)
        quality_report['outliers'] = outlier_report
        
        # Check for distribution shifts
        distribution_report = self.check_distribution_shift(current_data)
        quality_report['distribution_shift'] = distribution_report
        
        # Check for data type consistency
        type_report = self.check_data_types(current_data)
        quality_report['data_types'] = type_report
        
        # Overall quality score
        quality_score = self.calculate_quality_score(quality_report)
        quality_report['quality_score'] = quality_score
        
        return quality_report
    
    def check_missing_values(self, data: pd.DataFrame) -> Dict[str, Any]:
        """Check for missing values"""
        missing_counts = data.isnull().sum()
        missing_rates = missing_counts / len(data)
        
        issues = []
        for col, rate in missing_rates.items():
            if rate > self.quality_thresholds['missing_rate']:
                issues.append(f"Column {col} has {rate:.2%} missing values")
        
        return {
            'missing_counts': missing_counts.to_dict(),
            'missing_rates': missing_rates.to_dict(),
            'issues': issues
        }
    
    def check_outliers(self, data: pd.DataFrame) -> Dict[str, Any]:
        """Check for outliers using IQR method"""
        outlier_report = {}
        
        for col in data.select_dtypes(include=[np.number]).columns:
            if col in self.reference_data.columns:
                Q1 = self.reference_data[col].quantile(0.25)
                Q3 = self.reference_data[col].quantile(0.75)
                IQR = Q3 - Q1
                
                lower_bound = Q1 - 1.5 * IQR
                upper_bound = Q3 + 1.5 * IQR
                
                outliers = ((data[col] < lower_bound) | (data[col] > upper_bound)).sum()
                outlier_rate = outliers / len(data)
                
                outlier_report[col] = {
                    'outlier_count': outliers,
                    'outlier_rate': outlier_rate,
                    'is_anomalous': outlier_rate > self.quality_thresholds['outlier_rate']
                }
        
        return outlier_report
    
    def check_distribution_shift(self, data: pd.DataFrame) -> Dict[str, Any]:
        """Check for distribution shifts"""
        shift_report = {}
        
        for col in data.select_dtypes(include=[np.number]).columns:
            if col in self.reference_data.columns:
                # Kolmogorov-Smirnov test
                ks_stat, p_value = stats.ks_2samp(
                    self.reference_data[col].dropna(),
                    data[col].dropna()
                )
                
                # Calculate PSI
                psi = self.calculate_psi(
                    self.reference_data[col].dropna(),
                    data[col].dropna()
                )
                
                shift_report[col] = {
                    'ks_statistic': ks_stat,
                    'p_value': p_value,
                    'psi': psi,
                    'significant_shift': p_value < self.quality_thresholds['distribution_shift']
                }
        
        return shift_report
    
    def calculate_quality_score(self, quality_report: Dict[str, Any]) -> float:
        """Calculate overall data quality score"""
        score = 1.0
        
        # Penalize for missing values
        missing_issues = len(quality_report['missing_values']['issues'])
        score -= missing_issues * 0.1
        
        # Penalize for outliers
        outlier_issues = sum(1 for col_data in quality_report['outliers'].values() 
                           if col_data.get('is_anomalous', False))
        score -= outlier_issues * 0.05
        
        # Penalize for distribution shifts
        shift_issues = sum(1 for col_data in quality_report['distribution_shift'].values() 
                          if col_data.get('significant_shift', False))
        score -= shift_issues * 0.1
        
        return max(0.0, score)
```

### Alerting and Incident Response

#### 1. Intelligent Alerting

```python
from typing import Dict, Any, List
import logging
from datetime import datetime, timedelta

class AIAlertingSystem:
    def __init__(self, alert_rules: Dict[str, Any]):
        self.alert_rules = alert_rules
        self.alert_history = []
        self.logger = logging.getLogger(__name__)
    
    def check_alerts(self, metrics: Dict[str, Any]) -> List[Dict[str, Any]]:
        """Check metrics against alert rules"""
        alerts = []
        
        for rule_name, rule_config in self.alert_rules.items():
            if self.evaluate_rule(rule_config, metrics):
                alert = self.create_alert(rule_name, rule_config, metrics)
                alerts.append(alert)
        
        return alerts
    
    def evaluate_rule(self, rule_config: Dict[str, Any], metrics: Dict[str, Any]) -> bool:
        """Evaluate a single alert rule"""
        metric_name = rule_config['metric']
        threshold = rule_config['threshold']
        operator = rule_config['operator']
        
        if metric_name not in metrics:
            return False
        
        metric_value = metrics[metric_name]
        
        if operator == 'gt':
            return metric_value > threshold
        elif operator == 'lt':
            return metric_value < threshold
        elif operator == 'eq':
            return metric_value == threshold
        elif operator == 'gte':
            return metric_value >= threshold
        elif operator == 'lte':
            return metric_value <= threshold
        
        return False
    
    def create_alert(self, rule_name: str, rule_config: Dict[str, Any], 
                    metrics: Dict[str, Any]) -> Dict[str, Any]:
        """Create an alert"""
        alert = {
            'timestamp': datetime.now(),
            'rule_name': rule_name,
            'severity': rule_config.get('severity', 'warning'),
            'metric': rule_config['metric'],
            'threshold': rule_config['threshold'],
            'actual_value': metrics[rule_config['metric']],
            'message': f"{rule_config['metric']} {rule_config['operator']} {rule_config['threshold']}"
        }
        
        self.alert_history.append(alert)
        self.logger.warning(f"Alert triggered: {alert['message']}")
        
        return alert
    
    def get_alert_summary(self, hours: int = 24) -> Dict[str, Any]:
        """Get alert summary for the last N hours"""
        cutoff_time = datetime.now() - timedelta(hours=hours)
        recent_alerts = [alert for alert in self.alert_history 
                        if alert['timestamp'] > cutoff_time]
        
        return {
            'total_alerts': len(recent_alerts),
            'alerts_by_severity': self.group_alerts_by_severity(recent_alerts),
            'alerts_by_rule': self.group_alerts_by_rule(recent_alerts),
            'recent_alerts': recent_alerts[-10:]  # Last 10 alerts
        }
    
    def group_alerts_by_severity(self, alerts: List[Dict[str, Any]]) -> Dict[str, int]:
        """Group alerts by severity level"""
        severity_counts = {}
        for alert in alerts:
            severity = alert['severity']
            severity_counts[severity] = severity_counts.get(severity, 0) + 1
        return severity_counts
    
    def group_alerts_by_rule(self, alerts: List[Dict[str, Any]]) -> Dict[str, int]:
        """Group alerts by rule name"""
        rule_counts = {}
        for alert in alerts:
            rule = alert['rule_name']
            rule_counts[rule] = rule_counts.get(rule, 0) + 1
        return rule_counts
```

#### 2. Automated Incident Response

```python
class IncidentResponseSystem:
    def __init__(self, response_actions: Dict[str, Any]):
        self.response_actions = response_actions
        self.logger = logging.getLogger(__name__)
    
    def handle_alert(self, alert: Dict[str, Any]) -> Dict[str, Any]:
        """Handle an alert with automated response"""
        response = {
            'alert_id': alert.get('rule_name'),
            'timestamp': datetime.now(),
            'actions_taken': [],
            'status': 'handled'
        }
        
        # Determine response actions based on alert
        actions = self.determine_response_actions(alert)
        
        for action in actions:
            try:
                result = self.execute_action(action, alert)
                response['actions_taken'].append({
                    'action': action,
                    'result': result,
                    'status': 'success'
                })
            except Exception as e:
                response['actions_taken'].append({
                    'action': action,
                    'result': str(e),
                    'status': 'failed'
                })
                self.logger.error(f"Failed to execute action {action}: {e}")
        
        return response
    
    def determine_response_actions(self, alert: Dict[str, Any]) -> List[str]:
        """Determine appropriate response actions for an alert"""
        actions = []
        
        # High severity alerts get immediate response
        if alert['severity'] == 'critical':
            actions.extend(['scale_up', 'notify_team', 'create_incident'])
        elif alert['severity'] == 'warning':
            actions.append('log_issue')
        
        # Model-specific responses
        if 'model_drift' in alert['rule_name']:
            actions.append('trigger_retraining')
        elif 'data_quality' in alert['rule_name']:
            actions.append('validate_data_pipeline')
        
        return actions
    
    def execute_action(self, action: str, alert: Dict[str, Any]) -> str:
        """Execute a specific response action"""
        if action == 'scale_up':
            return self.scale_up_services()
        elif action == 'notify_team':
            return self.notify_team(alert)
        elif action == 'create_incident':
            return self.create_incident(alert)
        elif action == 'trigger_retraining':
            return self.trigger_model_retraining()
        elif action == 'validate_data_pipeline':
            return self.validate_data_pipeline()
        else:
            return f"Unknown action: {action}"
    
    def scale_up_services(self) -> str:
        """Scale up model serving services"""
        # Implementation would scale up Kubernetes deployments
        return "Scaled up model serving services"
    
    def notify_team(self, alert: Dict[str, Any]) -> str:
        """Send notification to team"""
        # Implementation would send Slack/email notifications
        return f"Notified team about {alert['rule_name']}"
    
    def create_incident(self, alert: Dict[str, Any]) -> str:
        """Create incident in tracking system"""
        # Implementation would create incident in PagerDuty/Jira
        return f"Created incident for {alert['rule_name']}"
    
    def trigger_model_retraining(self) -> str:
        """Trigger model retraining pipeline"""
        # Implementation would trigger retraining workflow
        return "Triggered model retraining pipeline"
    
    def validate_data_pipeline(self) -> str:
        """Validate data pipeline"""
        # Implementation would run data validation
        return "Validated data pipeline"
```

**In Agent Bruno**:
- **Comprehensive Monitoring**: Tracks model performance, data quality, and system metrics
- **Drift Detection**: Monitors for data and model drift
- **Explainability**: Provides SHAP and LIME explanations for predictions
- **Intelligent Alerting**: Smart alerting with automated response
- **Incident Response**: Automated incident handling and escalation
- **Observability Stack**: Integrates with Grafana, Prometheus, and Logfire

📖 **Related Docs**: [OBSERVABILITY.md](./OBSERVABILITY.md) | [PRODUCTION_READINESS.md](./PRODUCTION_READINESS.md) | [SLO_SETUP.md](./SLO_SETUP.md)

---

## 10. Event-Driven AI Architectures

### What are Event-Driven AI Architectures?

Event-driven AI architectures use events and messages to coordinate AI systems, enabling loose coupling, scalability, and real-time processing. These architectures are essential for building responsive, distributed AI systems.

**Key Benefits**:
- **Loose Coupling**: Components communicate through events, not direct calls
- **Scalability**: Easy to scale individual components independently
- **Real-time Processing**: Immediate response to events
- **Fault Tolerance**: System continues working even if some components fail
- **Flexibility**: Easy to add new components or modify existing ones

### Event-Driven Architecture Patterns

#### 1. Event Sourcing

```python
from dataclasses import dataclass
from typing import List, Any, Dict
from datetime import datetime
import json

@dataclass
class Event:
    event_id: str
    event_type: str
    timestamp: datetime
    data: Dict[str, Any]
    version: int = 1

class EventStore:
    def __init__(self, storage_backend):
        self.storage = storage_backend
        self.events = []
    
    def append_event(self, event: Event):
        """Append an event to the store"""
        self.events.append(event)
        self.storage.save_event(event)
    
    def get_events(self, entity_id: str, from_version: int = 0) -> List[Event]:
        """Get events for a specific entity"""
        return [event for event in self.events 
                if event.data.get('entity_id') == entity_id 
                and event.version >= from_version]
    
    def get_events_by_type(self, event_type: str) -> List[Event]:
        """Get events by type"""
        return [event for event in self.events if event.event_type == event_type]

class AIEventSourcing:
    def __init__(self, event_store: EventStore):
        self.event_store = event_store
    
    def handle_prediction_request(self, request_data: Dict[str, Any]) -> Event:
        """Handle a prediction request event"""
        event = Event(
            event_id=f"pred_{datetime.now().timestamp()}",
            event_type="prediction_requested",
            timestamp=datetime.now(),
            data=request_data
        )
        self.event_store.append_event(event)
        return event
    
    def handle_prediction_completed(self, prediction_id: str, result: Dict[str, Any]) -> Event:
        """Handle a prediction completion event"""
        event = Event(
            event_id=f"pred_complete_{datetime.now().timestamp()}",
            event_type="prediction_completed",
            timestamp=datetime.now(),
            data={
                'prediction_id': prediction_id,
                'result': result
            }
        )
        self.event_store.append_event(event)
        return event
```

#### 2. CQRS (Command Query Responsibility Segregation)

```python
from abc import ABC, abstractmethod
from typing import Dict, Any, List
import asyncio

class Command:
    def __init__(self, command_type: str, data: Dict[str, Any]):
        self.command_type = command_type
        self.data = data
        self.timestamp = datetime.now()

class Query:
    def __init__(self, query_type: str, filters: Dict[str, Any]):
        self.query_type = query_type
        self.filters = filters
        self.timestamp = datetime.now()

class CommandHandler(ABC):
    @abstractmethod
    async def handle(self, command: Command) -> Any:
        pass

class QueryHandler(ABC):
    @abstractmethod
    async def handle(self, query: Query) -> Any:
        pass

class AICommandHandler(CommandHandler):
    def __init__(self, model_service, event_store):
        self.model_service = model_service
        self.event_store = event_store
    
    async def handle(self, command: Command) -> Any:
        if command.command_type == "train_model":
            return await self.handle_train_model(command)
        elif command.command_type == "make_prediction":
            return await self.handle_make_prediction(command)
        else:
            raise ValueError(f"Unknown command type: {command.command_type}")
    
    async def handle_train_model(self, command: Command) -> str:
        """Handle model training command"""
        model_id = await self.model_service.train_model(command.data)
        
        # Emit event
        event = Event(
            event_id=f"train_{datetime.now().timestamp()}",
            event_type="model_trained",
            timestamp=datetime.now(),
            data={'model_id': model_id, 'training_data': command.data}
        )
        self.event_store.append_event(event)
        
        return model_id
    
    async def handle_make_prediction(self, command: Command) -> Dict[str, Any]:
        """Handle prediction command"""
        prediction = await self.model_service.predict(command.data)
        
        # Emit event
        event = Event(
            event_id=f"pred_{datetime.now().timestamp()}",
            event_type="prediction_made",
            timestamp=datetime.now(),
            data={'prediction': prediction, 'input': command.data}
        )
        self.event_store.append_event(event)
        
        return prediction

class AIQueryHandler(QueryHandler):
    def __init__(self, read_model):
        self.read_model = read_model
    
    async def handle(self, query: Query) -> Any:
        if query.query_type == "get_model_performance":
            return await self.get_model_performance(query.filters)
        elif query.query_type == "get_prediction_history":
            return await self.get_prediction_history(query.filters)
        else:
            raise ValueError(f"Unknown query type: {query.query_type}")
    
    async def get_model_performance(self, filters: Dict[str, Any]) -> Dict[str, Any]:
        """Get model performance metrics"""
        return await self.read_model.get_performance_metrics(filters)
    
    async def get_prediction_history(self, filters: Dict[str, Any]) -> List[Dict[str, Any]]:
        """Get prediction history"""
        return await self.read_model.get_prediction_history(filters)
```

### Message Brokers and Event Streaming

#### 1. Apache Kafka Integration

```python
from kafka import KafkaProducer, KafkaConsumer
import json
from typing import Dict, Any, Callable
import threading

class KafkaEventStream:
    def __init__(self, bootstrap_servers: List[str]):
        self.bootstrap_servers = bootstrap_servers
        self.producer = KafkaProducer(
            bootstrap_servers=bootstrap_servers,
            value_serializer=lambda v: json.dumps(v).encode('utf-8')
        )
        self.consumers = {}
    
    def publish_event(self, topic: str, event: Dict[str, Any]):
        """Publish an event to a topic"""
        self.producer.send(topic, event)
        self.producer.flush()
    
    def subscribe_to_topic(self, topic: str, handler: Callable[[Dict[str, Any]], None]):
        """Subscribe to a topic with a handler"""
        consumer = KafkaConsumer(
            topic,
            bootstrap_servers=self.bootstrap_servers,
            value_deserializer=lambda m: json.loads(m.decode('utf-8')),
            group_id=f"ai_consumer_{topic}"
        )
        
        def consume_messages():
            for message in consumer:
                try:
                    handler(message.value)
                except Exception as e:
                    print(f"Error processing message: {e}")
        
        thread = threading.Thread(target=consume_messages)
        thread.daemon = True
        thread.start()
        
        self.consumers[topic] = consumer
        return consumer

class AIEventProcessor:
    def __init__(self, event_stream: KafkaEventStream):
        self.event_stream = event_stream
        self.setup_event_handlers()
    
    def setup_event_handlers(self):
        """Setup event handlers for different topics"""
        # Model training events
        self.event_stream.subscribe_to_topic(
            "model_training_requests",
            self.handle_training_request
        )
        
        # Prediction events
        self.event_stream.subscribe_to_topic(
            "prediction_requests",
            self.handle_prediction_request
        )
        
        # Monitoring events
        self.event_stream.subscribe_to_topic(
            "model_monitoring",
            self.handle_monitoring_event
        )
    
    def handle_training_request(self, event: Dict[str, Any]):
        """Handle model training request"""
        print(f"Processing training request: {event['model_id']}")
        # Implement training logic
        self.event_stream.publish_event("model_training_completed", {
            "model_id": event['model_id'],
            "status": "completed",
            "metrics": {"accuracy": 0.95}
        })
    
    def handle_prediction_request(self, event: Dict[str, Any]):
        """Handle prediction request"""
        print(f"Processing prediction request: {event['request_id']}")
        # Implement prediction logic
        self.event_stream.publish_event("prediction_completed", {
            "request_id": event['request_id'],
            "prediction": {"class": "positive", "confidence": 0.87}
        })
    
    def handle_monitoring_event(self, event: Dict[str, Any]):
        """Handle monitoring event"""
        print(f"Processing monitoring event: {event['metric']}")
        # Implement monitoring logic
        if event['metric'] == 'model_drift':
            self.event_stream.publish_event("retraining_triggered", {
                "model_id": event['model_id'],
                "reason": "drift_detected"
            })
```

#### 2. RabbitMQ Integration

```python
import pika
import json
from typing import Dict, Any, Callable

class RabbitMQEventBroker:
    def __init__(self, connection_url: str):
        self.connection_url = connection_url
        self.connection = pika.BlockingConnection(pika.URLParameters(connection_url))
        self.channel = self.connection.channel()
    
    def declare_exchange(self, exchange_name: str, exchange_type: str = 'topic'):
        """Declare an exchange"""
        self.channel.exchange_declare(
            exchange=exchange_name,
            exchange_type=exchange_type,
            durable=True
        )
    
    def publish_message(self, exchange: str, routing_key: str, message: Dict[str, Any]):
        """Publish a message to an exchange"""
        self.channel.basic_publish(
            exchange=exchange,
            routing_key=routing_key,
            body=json.dumps(message),
            properties=pika.BasicProperties(delivery_mode=2)  # Make message persistent
        )
    
    def consume_messages(self, queue_name: str, handler: Callable[[Dict[str, Any]], None]):
        """Consume messages from a queue"""
        def callback(ch, method, properties, body):
            try:
                message = json.loads(body)
                handler(message)
                ch.basic_ack(delivery_tag=method.delivery_tag)
            except Exception as e:
                print(f"Error processing message: {e}")
                ch.basic_nack(delivery_tag=method.delivery_tag, requeue=True)
        
        self.channel.basic_consume(
            queue=queue_name,
            on_message_callback=callback
        )
        
        self.channel.start_consuming()

class AIEventRouter:
    def __init__(self, broker: RabbitMQEventBroker):
        self.broker = broker
        self.setup_exchanges_and_queues()
    
    def setup_exchanges_and_queues(self):
        """Setup exchanges and queues for AI events"""
        # Declare exchanges
        self.broker.declare_exchange('ai_events', 'topic')
        self.broker.declare_exchange('model_events', 'topic')
        self.broker.declare_exchange('prediction_events', 'topic')
        
        # Setup queues
        self.broker.channel.queue_declare('model_training_queue', durable=True)
        self.broker.channel.queue_declare('prediction_queue', durable=True)
        self.broker.channel.queue_declare('monitoring_queue', durable=True)
        
        # Bind queues to exchanges
        self.broker.channel.queue_bind(
            exchange='model_events',
            queue='model_training_queue',
            routing_key='model.training.*'
        )
        
        self.broker.channel.queue_bind(
            exchange='prediction_events',
            queue='prediction_queue',
            routing_key='prediction.*'
        )
        
        self.broker.channel.queue_bind(
            exchange='ai_events',
            queue='monitoring_queue',
            routing_key='monitoring.*'
        )
    
    def route_training_event(self, event: Dict[str, Any]):
        """Route model training event"""
        self.broker.publish_message(
            'model_events',
            'model.training.requested',
            event
        )
    
    def route_prediction_event(self, event: Dict[str, Any]):
        """Route prediction event"""
        self.broker.publish_message(
            'prediction_events',
            'prediction.requested',
            event
        )
    
    def route_monitoring_event(self, event: Dict[str, Any]):
        """Route monitoring event"""
        self.broker.publish_message(
            'ai_events',
            'monitoring.alert',
            event
        )
```

### CloudEvents Integration

#### 1. CloudEvents for AI Systems

```python
from cloudevents.http import CloudEvent
from cloudevents.conversion import to_structured
import json
from typing import Dict, Any

class AICloudEvent:
    def __init__(self, event_type: str, source: str, data: Dict[str, Any]):
        self.event = CloudEvent({
            "type": event_type,
            "source": source,
            "datacontenttype": "application/json"
        }, data)
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert CloudEvent to dictionary"""
        return {
            "id": self.event["id"],
            "type": self.event["type"],
            "source": self.event["source"],
            "time": self.event["time"],
            "data": self.event.data
        }
    
    def to_json(self) -> str:
        """Convert CloudEvent to JSON"""
        return json.dumps(self.to_dict())

class AIEventPublisher:
    def __init__(self, event_broker):
        self.broker = event_broker
    
    def publish_model_trained_event(self, model_id: str, metrics: Dict[str, Any]):
        """Publish model trained event"""
        event = AICloudEvent(
            event_type="ai.model.trained",
            source="agent-bruno/model-service",
            data={
                "model_id": model_id,
                "metrics": metrics,
                "timestamp": datetime.now().isoformat()
            }
        )
        self.broker.publish_event("ai_events", event.to_dict())
    
    def publish_prediction_made_event(self, request_id: str, prediction: Dict[str, Any]):
        """Publish prediction made event"""
        event = AICloudEvent(
            event_type="ai.prediction.made",
            source="agent-bruno/prediction-service",
            data={
                "request_id": request_id,
                "prediction": prediction,
                "timestamp": datetime.now().isoformat()
            }
        )
        self.broker.publish_event("ai_events", event.to_dict())
    
    def publish_drift_detected_event(self, model_id: str, drift_metrics: Dict[str, Any]):
        """Publish drift detected event"""
        event = AICloudEvent(
            event_type="ai.drift.detected",
            source="agent-bruno/monitoring-service",
            data={
                "model_id": model_id,
                "drift_metrics": drift_metrics,
                "timestamp": datetime.now().isoformat()
            }
        )
        self.broker.publish_event("ai_events", event.to_dict())

class AIEventConsumer:
    def __init__(self, event_broker):
        self.broker = event_broker
        self.setup_consumers()
    
    def setup_consumers(self):
        """Setup event consumers"""
        self.broker.subscribe_to_topic("ai_events", self.handle_ai_event)
        self.broker.subscribe_to_topic("model_events", self.handle_model_event)
        self.broker.subscribe_to_topic("prediction_events", self.handle_prediction_event)
    
    def handle_ai_event(self, event: Dict[str, Any]):
        """Handle general AI events"""
        event_type = event.get("type", "")
        
        if event_type == "ai.model.trained":
            self.handle_model_trained(event)
        elif event_type == "ai.prediction.made":
            self.handle_prediction_made(event)
        elif event_type == "ai.drift.detected":
            self.handle_drift_detected(event)
    
    def handle_model_trained(self, event: Dict[str, Any]):
        """Handle model trained event"""
        model_id = event["data"]["model_id"]
        metrics = event["data"]["metrics"]
        print(f"Model {model_id} trained with metrics: {metrics}")
        
        # Update model registry
        # Trigger model validation
        # Update monitoring dashboards
    
    def handle_prediction_made(self, event: Dict[str, Any]):
        """Handle prediction made event"""
        request_id = event["data"]["request_id"]
        prediction = event["data"]["prediction"]
        print(f"Prediction {request_id} completed: {prediction}")
        
        # Log prediction for monitoring
        # Update performance metrics
        # Store in prediction history
    
    def handle_drift_detected(self, event: Dict[str, Any]):
        """Handle drift detected event"""
        model_id = event["data"]["model_id"]
        drift_metrics = event["data"]["drift_metrics"]
        print(f"Drift detected for model {model_id}: {drift_metrics}")
        
        # Trigger model retraining
        # Send alerts to team
        # Update monitoring dashboards
```

### Event-Driven AI Workflows

#### 1. Automated Model Retraining

```python
class AutomatedRetrainingWorkflow:
    def __init__(self, event_broker, model_service, data_service):
        self.broker = event_broker
        self.model_service = model_service
        self.data_service = data_service
        self.setup_workflow()
    
    def setup_workflow(self):
        """Setup automated retraining workflow"""
        self.broker.subscribe_to_topic("ai.drift.detected", self.handle_drift_detected)
        self.broker.subscribe_to_topic("ai.performance.degraded", self.handle_performance_degraded)
        self.broker.subscribe_to_topic("ai.data.updated", self.handle_data_updated)
    
    def handle_drift_detected(self, event: Dict[str, Any]):
        """Handle drift detected event"""
        model_id = event["data"]["model_id"]
        drift_metrics = event["data"]["drift_metrics"]
        
        print(f"Drift detected for model {model_id}, triggering retraining")
        
        # Start retraining process
        retraining_id = self.model_service.start_retraining(model_id)
        
        # Publish retraining started event
        self.broker.publish_event("ai.retraining.started", {
            "model_id": model_id,
            "retraining_id": retraining_id,
            "reason": "drift_detected",
            "drift_metrics": drift_metrics
        })
    
    def handle_performance_degraded(self, event: Dict[str, Any]):
        """Handle performance degraded event"""
        model_id = event["data"]["model_id"]
        performance_metrics = event["data"]["performance_metrics"]
        
        print(f"Performance degraded for model {model_id}, triggering retraining")
        
        # Start retraining process
        retraining_id = self.model_service.start_retraining(model_id)
        
        # Publish retraining started event
        self.broker.publish_event("ai.retraining.started", {
            "model_id": model_id,
            "retraining_id": retraining_id,
            "reason": "performance_degraded",
            "performance_metrics": performance_metrics
        })
    
    def handle_data_updated(self, event: Dict[str, Any]):
        """Handle data updated event"""
        data_source = event["data"]["data_source"]
        data_size = event["data"]["data_size"]
        
        print(f"Data updated from {data_source}, size: {data_size}")
        
        # Check if retraining is needed
        if self.should_retrain(data_source, data_size):
            # Get latest model
            latest_model = self.model_service.get_latest_model()
            
            # Start retraining
            retraining_id = self.model_service.start_retraining(latest_model.id)
            
            # Publish retraining started event
            self.broker.publish_event("ai.retraining.started", {
                "model_id": latest_model.id,
                "retraining_id": retraining_id,
                "reason": "data_updated",
                "data_source": data_source,
                "data_size": data_size
            })
    
    def should_retrain(self, data_source: str, data_size: int) -> bool:
        """Determine if retraining is needed based on data updates"""
        # Implement logic to determine if retraining is needed
        # This could be based on data size, time since last training, etc.
        return data_size > 1000  # Example threshold
```

#### 2. Real-time Model Serving

```python
class RealTimeModelServing:
    def __init__(self, event_broker, model_registry, prediction_service):
        self.broker = event_broker
        self.model_registry = model_registry
        self.prediction_service = prediction_service
        self.setup_serving()
    
    def setup_serving(self):
        """Setup real-time model serving"""
        self.broker.subscribe_to_topic("prediction.requests", self.handle_prediction_request)
        self.broker.subscribe_to_topic("model.updated", self.handle_model_updated)
        self.broker.subscribe_to_topic("model.performance.alert", self.handle_performance_alert)
    
    def handle_prediction_request(self, event: Dict[str, Any]):
        """Handle prediction request"""
        request_id = event["data"]["request_id"]
        input_data = event["data"]["input_data"]
        model_id = event["data"].get("model_id")
        
        try:
            # Get model (use latest if not specified)
            if not model_id:
                model = self.model_registry.get_latest_model()
            else:
                model = self.model_registry.get_model(model_id)
            
            # Make prediction
            prediction = self.prediction_service.predict(model, input_data)
            
            # Publish prediction completed event
            self.broker.publish_event("prediction.completed", {
                "request_id": request_id,
                "model_id": model.id,
                "prediction": prediction,
                "timestamp": datetime.now().isoformat()
            })
            
        except Exception as e:
            # Publish prediction failed event
            self.broker.publish_event("prediction.failed", {
                "request_id": request_id,
                "error": str(e),
                "timestamp": datetime.now().isoformat()
            })
    
    def handle_model_updated(self, event: Dict[str, Any]):
        """Handle model updated event"""
        model_id = event["data"]["model_id"]
        model_version = event["data"]["model_version"]
        
        print(f"Model {model_id} updated to version {model_version}")
        
        # Update model in serving system
        self.prediction_service.update_model(model_id, model_version)
        
        # Publish model serving updated event
        self.broker.publish_event("model.serving.updated", {
            "model_id": model_id,
            "model_version": model_version,
            "timestamp": datetime.now().isoformat()
        })
    
    def handle_performance_alert(self, event: Dict[str, Any]):
        """Handle performance alert"""
        model_id = event["data"]["model_id"]
        performance_metrics = event["data"]["performance_metrics"]
        
        print(f"Performance alert for model {model_id}: {performance_metrics}")
        
        # Take corrective action (e.g., scale up, switch model)
        if performance_metrics["latency"] > 1000:  # 1 second threshold
            self.prediction_service.scale_up_model(model_id)
        
        # Publish corrective action event
        self.broker.publish_event("model.serving.corrective_action", {
            "model_id": model_id,
            "action": "scaled_up",
            "reason": "high_latency",
            "performance_metrics": performance_metrics
        })
```

**In Agent Bruno**:
- **Event-Driven Architecture**: Uses CloudEvents and RabbitMQ for event coordination
- **Real-time Processing**: Immediate response to user queries and system events
- **Automated Workflows**: Self-healing and self-improving AI system
- **Loose Coupling**: Components communicate through events, not direct calls
- **Scalability**: Easy to scale individual components independently
- **Fault Tolerance**: System continues working even if some components fail

📖 **Related Docs**: [ARCHITECTURE.md](./ARCHITECTURE.md) | [MCP_WORKFLOWS.md](./MCP_WORKFLOWS.md) | [OBSERVABILITY.md](./OBSERVABILITY.md)