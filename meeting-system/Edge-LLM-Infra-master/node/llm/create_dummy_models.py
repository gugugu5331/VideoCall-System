#!/usr/bin/env python3
"""
Create dummy ONNX models for testing
"""

import numpy as np
import sys

try:
    import onnx
    from onnx import helper, TensorProto
except ImportError:
    print("Installing onnx...")
    import subprocess
    subprocess.check_call([sys.executable, "-m", "pip", "install", "onnx"])
    import onnx
    from onnx import helper, TensorProto

def create_dummy_asr_model(output_path):
    """Create a dummy ASR model"""
    print(f"Creating dummy ASR model: {output_path}")
    
    # Input: audio features [batch_size, sequence_length, feature_dim]
    input_tensor = helper.make_tensor_value_info('audio_input', TensorProto.FLOAT, [1, 100, 80])
    
    # Output: logits [batch_size, sequence_length, vocab_size]
    output_tensor = helper.make_tensor_value_info('transcription_output', TensorProto.FLOAT, [1, 100, 1000])
    
    # Create a simple identity-like operation
    node = helper.make_node(
        'Identity',
        inputs=['audio_input'],
        outputs=['transcription_output']
    )
    
    # Create the graph
    graph = helper.make_graph(
        [node],
        'asr_model',
        [input_tensor],
        [output_tensor]
    )
    
    # Create the model with IR version 9 (compatible with ONNX Runtime 1.16.3)
    model = helper.make_model(graph, producer_name='dummy_asr', ir_version=9, opset_imports=[helper.make_opsetid("", 13)])

    # Save the model
    onnx.save(model, output_path)
    print(f"✓ ASR model saved to {output_path}")

def create_dummy_emotion_model(output_path):
    """Create a dummy emotion detection model"""
    print(f"Creating dummy Emotion model: {output_path}")
    
    # Input: text embeddings [batch_size, sequence_length, embedding_dim]
    input_tensor = helper.make_tensor_value_info('text_input', TensorProto.FLOAT, [1, 128, 768])
    
    # Output: emotion logits [batch_size, num_emotions]
    output_tensor = helper.make_tensor_value_info('emotion_output', TensorProto.FLOAT, [1, 7])
    
    # Create a simple reduce mean + identity operation
    node1 = helper.make_node(
        'ReduceMean',
        inputs=['text_input'],
        outputs=['reduced'],
        axes=[1, 2]
    )
    
    node2 = helper.make_node(
        'Identity',
        inputs=['reduced'],
        outputs=['emotion_output']
    )
    
    # Create the graph
    graph = helper.make_graph(
        [node1, node2],
        'emotion_model',
        [input_tensor],
        [output_tensor]
    )
    
    # Create the model with IR version 9
    model = helper.make_model(graph, producer_name='dummy_emotion', ir_version=9, opset_imports=[helper.make_opsetid("", 13)])

    # Save the model
    onnx.save(model, output_path)
    print(f"✓ Emotion model saved to {output_path}")

def create_dummy_synthesis_model(output_path):
    """Create a dummy synthesis detection model"""
    print(f"Creating dummy Synthesis Detection model: {output_path}")
    
    # Input: audio features [batch_size, sequence_length, feature_dim]
    input_tensor = helper.make_tensor_value_info('audio_input', TensorProto.FLOAT, [1, 100, 80])
    
    # Output: binary classification [batch_size, 1]
    output_tensor = helper.make_tensor_value_info('synthesis_output', TensorProto.FLOAT, [1, 1])
    
    # Create a simple reduce mean operation
    node = helper.make_node(
        'ReduceMean',
        inputs=['audio_input'],
        outputs=['synthesis_output'],
        axes=[1, 2]
    )
    
    # Create the graph
    graph = helper.make_graph(
        [node],
        'synthesis_model',
        [input_tensor],
        [output_tensor]
    )
    
    # Create the model with IR version 9
    model = helper.make_model(graph, producer_name='dummy_synthesis', ir_version=9, opset_imports=[helper.make_opsetid("", 13)])

    # Save the model
    onnx.save(model, output_path)
    print(f"✓ Synthesis Detection model saved to {output_path}")

def main():
    import os
    
    # Create models directory
    models_dir = "/work/models" if os.path.exists("/work") else "./models"
    os.makedirs(models_dir, exist_ok=True)
    
    print("=" * 60)
    print("Creating Dummy ONNX Models for Testing")
    print("=" * 60)
    print(f"Models directory: {models_dir}\n")
    
    # Create models
    create_dummy_asr_model(os.path.join(models_dir, "asr-model.onnx"))
    create_dummy_emotion_model(os.path.join(models_dir, "emotion-model.onnx"))
    create_dummy_synthesis_model(os.path.join(models_dir, "synthesis-model.onnx"))
    
    print("\n" + "=" * 60)
    print("✓ All dummy models created successfully!")
    print("=" * 60)

if __name__ == "__main__":
    main()

