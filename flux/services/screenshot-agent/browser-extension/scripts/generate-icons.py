#!/usr/bin/env python3
"""
Gera ícones simples para a extensão Screenshot Agent
"""
from PIL import Image, ImageDraw
import os

def create_icon(size, output_path):
    """Cria um ícone simples com uma câmera/screenshot"""
    # Criar imagem com fundo transparente
    img = Image.new('RGBA', (size, size), (0, 0, 0, 0))
    draw = ImageDraw.Draw(img)
    
    # Cores
    bg_color = (21, 93, 220, 255)  # #175DDC
    white = (255, 255, 255, 255)
    
    # Desenhar fundo arredondado
    margin = size // 8
    draw.rounded_rectangle(
        [margin, margin, size - margin, size - margin],
        radius=size // 6,
        fill=bg_color
    )
    
    # Desenhar câmera simples (retângulo com círculo no centro)
    camera_width = size // 2
    camera_height = size // 2.5
    camera_x = (size - camera_width) // 2
    camera_y = (size - camera_height) // 2
    
    # Corpo da câmera
    draw.rounded_rectangle(
        [camera_x, camera_y, camera_x + camera_width, camera_y + camera_height],
        radius=size // 20,
        fill=white,
        outline=white,
        width=size // 40
    )
    
    # Lente (círculo)
    lens_size = size // 4
    lens_x = size // 2 - lens_size // 2
    lens_y = size // 2 - lens_size // 2
    draw.ellipse(
        [lens_x, lens_y, lens_x + lens_size, lens_y + lens_size],
        fill=bg_color,
        outline=white,
        width=size // 40
    )
    
    # Flash (pequeno círculo no canto)
    flash_size = size // 8
    flash_x = camera_x + camera_width - flash_size - size // 16
    flash_y = camera_y + size // 16
    draw.ellipse(
        [flash_x, flash_y, flash_x + flash_size, flash_y + flash_size],
        fill=white
    )
    
    # Salvar
    img.save(output_path, 'PNG')
    print(f"✓ Criado: {output_path}")

def main():
    script_dir = os.path.dirname(os.path.abspath(__file__))
    chrome_icons_dir = os.path.join(script_dir, '..', 'chrome', 'icons')
    safari_icons_dir = os.path.join(script_dir, '..', 'safari', 'icons')
    
    # Criar diretórios
    os.makedirs(chrome_icons_dir, exist_ok=True)
    os.makedirs(safari_icons_dir, exist_ok=True)
    
    # Gerar ícones para Chrome
    sizes = [16, 48, 128]
    for size in sizes:
        create_icon(size, os.path.join(chrome_icons_dir, f'icon{size}.png'))
    
    # Gerar ícones para Safari (mesmos tamanhos)
    for size in sizes:
        create_icon(size, os.path.join(safari_icons_dir, f'icon{size}.png'))
    
    print("\n✅ Todos os ícones foram gerados!")

if __name__ == '__main__':
    try:
        from PIL import Image, ImageDraw
        main()
    except ImportError:
        print("⚠️  PIL/Pillow não está instalado. Instalando...")
        import subprocess
        import sys
        subprocess.check_call([sys.executable, '-m', 'pip', 'install', 'Pillow'])
        from PIL import Image, ImageDraw
        main()
