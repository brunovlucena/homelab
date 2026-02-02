# 📸 Screenshot Agent - Browser Extension

Extensão de navegador para Chrome e Safari que captura screenshots da tela e envia para um agente no homelab para análise.

## 🎯 Funcionalidades

- ✅ Captura screenshots da aba atual
- ✅ Envia screenshots para agente no homelab
- ✅ Configurável (URL do agente, formato de imagem)
- ✅ Interface simples e intuitiva
- ✅ Suporte para Chrome e Safari

## 📁 Estrutura do Projeto

```
browser-extension/
├── chrome/              # Extensão Chrome
│   ├── manifest.json
│   ├── background.js
│   ├── popup.html
│   ├── popup.js
│   ├── options.html
│   ├── options.js
│   └── icons/
├── safari/              # Extensão Safari
│   ├── manifest.json
│   ├── background.js
│   ├── popup.html
│   ├── popup.js
│   ├── options.html
│   ├── options.js
│   └── icons/
├── shared/              # Código compartilhado
│   └── config.js
└── scripts/             # Scripts auxiliares
    └── generate-icons.py
```

## 🚀 Instalação

### Chrome

1. **Preparar a extensão:**
   ```bash
   cd browser-extension/chrome
   ```

2. **Gerar ícones (se necessário):**
   ```bash
   cd ../scripts
   python3 generate-icons.py
   ```

3. **Carregar no Chrome:**
   - Abra o Chrome e vá para `chrome://extensions/`
   - Ative o "Modo do desenvolvedor" (Developer mode)
   - Clique em "Carregar sem compactação" (Load unpacked)
   - Selecione o diretório `chrome/`

4. **Configurar:**
   - Clique no ícone da extensão
   - Clique em "⚙️ Configurações"
   - Configure a URL do agente (ex: `http://localhost:8080/api/v1/screenshots`)
   - Salve as configurações

### Safari

1. **Preparar a extensão:**
   ```bash
   cd browser-extension/safari
   ```

2. **Nota sobre Safari:**
   - Safari requer um projeto Xcode para desenvolvimento
   - Para desenvolvimento, use Safari Web Extension Converter
   - Ou use o Safari Technology Preview que suporta Web Extensions diretamente

3. **Usando Safari Technology Preview:**
   - Abra o Safari Technology Preview
   - Vá em Preferences → Extensions
   - Ative "Allow Unsigned Extensions"
   - Arraste o diretório `safari/` para a área de extensões

4. **Para produção:**
   - Use o Safari Web Extension Converter (Xcode)
   - Ou compile usando o Xcode Project Generator

## ⚙️ Configuração

### URL do Agente

Configure a URL do endpoint que receberá os screenshots. Exemplos:

- **Local:** `http://localhost:8080/api/v1/screenshots`
- **Homelab (local):** `http://homelab-api.local:8080/api/v1/screenshots`
- **Homelab (cloud):** `https://api.lucena.cloud/api/v1/screenshots`
- **Kubernetes (interno):** `http://mobile-api.homelab-services.svc.cluster.local:8080/api/v1/screenshots`

### Formato de Imagem

- **PNG** (recomendado) - Melhor qualidade, sem compressão
- **JPEG** - Menor tamanho, com compressão

## 🔌 API do Backend

A extensão espera um endpoint POST que recebe um FormData com:

- `screenshot`: Arquivo de imagem (PNG ou JPEG)
- `url`: URL da página capturada
- `title`: Título da página
- `timestamp`: Timestamp ISO 8601

### Exemplo de Handler (Go/Gin)

```go
func handleScreenshot(c *gin.Context) {
    // Receber arquivo
    file, err := c.FormFile("screenshot")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Receber metadados
    url := c.PostForm("url")
    title := c.PostForm("title")
    timestamp := c.PostForm("timestamp")
    
    // Processar screenshot (salvar, analisar, etc.)
    // Aqui você pode:
    // - Salvar em MinIO/S3
    // - Enviar para agente de análise (GPT-4V, Claude, etc.)
    // - Processar com OCR
    // - etc.
    
    c.JSON(http.StatusOK, gin.H{
        "status": "success",
        "url": url,
        "title": title,
        "timestamp": timestamp,
        "message": "Screenshot recebido e processado"
    })
}
```

### Endpoint Recomendado

Adicione ao seu `mobile-api/main.go`:

```go
api.POST("/screenshots", handleScreenshot)
```

## 📝 Uso

1. **Capturar Screenshot:**
   - Navegue até a página desejada
   - Clique no ícone da extensão na barra de ferramentas
   - Clique em "📸 Capturar Screenshot"
   - Aguarde a confirmação

2. **Verificar Status:**
   - O popup mostra o status da captura
   - Mensagens de sucesso/erro são exibidas
   - Informações da página são mostradas

## 🔧 Desenvolvimento

### Gerar Ícones

```bash
cd scripts
python3 generate-icons.py
```

Requer: `Pillow` (instalado automaticamente se não disponível)

### Estrutura de Arquivos

- **manifest.json**: Configuração da extensão (permissões, ícones, etc.)
- **background.js**: Service worker que gerencia captura e upload
- **popup.html/js**: Interface do usuário
- **options.html/js**: Página de configurações

## 🐛 Troubleshooting

### Chrome

- **Erro de CORS:** Verifique se o backend tem CORS habilitado
- **Upload falha:** Verifique a URL do agente nas configurações
- **Screenshot não captura:** Verifique permissões da extensão

### Safari

- **Extensão não carrega:** Use Safari Technology Preview ou compile com Xcode
- **API não disponível:** Safari pode ter limitações na API de screenshots
- **Upload falha:** Verifique permissões de rede no Safari

## 📚 Referências

- [Chrome Extensions Documentation](https://developer.chrome.com/docs/extensions/)
- [Safari Web Extensions](https://developer.apple.com/documentation/safariservices/safari_web_extensions)
- [Manifest V3](https://developer.chrome.com/docs/extensions/mv3/intro/)

## 🔒 Segurança

- A extensão requer permissões para capturar screenshots
- Screenshots são enviados apenas para o URL configurado
- Configure HTTPS em produção
- Considere autenticação no endpoint do agente

## 📄 Licença

Este projeto faz parte do homelab pessoal.
