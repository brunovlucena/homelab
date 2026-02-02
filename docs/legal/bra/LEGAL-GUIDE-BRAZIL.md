# 🇧🇷 Guia Legal - Comercialização do Homelab no Brasil

> **Versão do Documento**: 1.0  
> **Última Atualização**: 11 de Dezembro de 2025  
> **Autor**: Bruno Lucena  
> **Jurisdição**: República Federativa do Brasil

---

## Sumário Executivo

Este documento fornece análise completa dos requisitos legais para comercializar a plataforma Homelab no Brasil, incluindo:

- Proteção de software e direitos autorais
- Conformidade com LGPD (Lei Geral de Proteção de Dados)
- Formação de pessoa jurídica
- Registro de marca e propriedade intelectual
- Licenciamento de software open source

### ⚠️ Pontos Críticos

| Risco | Item | Ação Necessária |
|-------|------|-----------------|
| 🔴 ALTO | slither-analyzer (AGPLv3) | Obter licença comercial ou substituir |
| 🔴 ALTO | Conformidade LGPD | Implementar antes de processar dados |
| 🟡 MÉDIO | Registro de Software no INPI | Recomendado para prova de autoria |
| 🟡 MÉDIO | Registro de Marca | Proteger nome "Homelab" |

---

## Índice

1. [Legislação Aplicável](#1-legislação-aplicável)
2. [Proteção de Software](#2-proteção-de-software)
3. [Conformidade LGPD](#3-conformidade-lgpd)
4. [Formação de Empresa](#4-formação-de-empresa)
5. [Propriedade Intelectual](#5-propriedade-intelectual)
6. [Licenças Open Source](#6-licenças-open-source)
7. [Tributação](#7-tributação)
8. [Checklist de Conformidade](#8-checklist-de-conformidade)

---

## 1. Legislação Aplicável

### Principais Leis

| Lei | Número | Assunto |
|-----|--------|---------|
| Lei do Software | 9.609/1998 | Proteção de programas de computador |
| Lei de Direitos Autorais | 9.610/1998 | Direitos autorais (subsidiária) |
| LGPD | 13.709/2018 | Proteção de dados pessoais |
| Marco Civil da Internet | 12.965/2014 | Regulamentação da internet |
| Código Civil | 10.406/2002 | Contratos e obrigações |
| Lei da Propriedade Industrial | 9.279/1996 | Marcas e patentes |

### Órgãos Reguladores

| Órgão | Função |
|-------|--------|
| **INPI** | Instituto Nacional da Propriedade Industrial - registro de marcas e software |
| **ANPD** | Autoridade Nacional de Proteção de Dados - fiscalização LGPD |
| **Receita Federal** | Tributação e registro de empresas |
| **Juntas Comerciais** | Registro de pessoas jurídicas |

---

## 2. Proteção de Software

### 2.1 Direitos Autorais sobre Software

No Brasil, software é protegido por **direitos autorais**, não por patentes.

| Característica | Descrição |
|----------------|-----------|
| **Proteção Automática** | Nasce com a criação, sem necessidade de registro |
| **Prazo de Proteção** | 50 anos a partir de 1º de janeiro do ano seguinte à publicação |
| **O que é protegido** | Expressão literal (código-fonte), não funcionalidades |
| **O que NÃO é protegido** | Ideias, algoritmos abstratos, funcionalidades |

### 2.2 Registro de Software no INPI

**Por que registrar?**
- Prova de autoria em disputas judiciais
- Evidência da data de criação
- Facilita licenciamento e venda
- Requisito para alguns editais públicos

**Processo de Registro (100% eletrônico)**

| Etapa | Descrição | Prazo |
|-------|-----------|-------|
| 1. Cadastro | Criar conta no e-INPI | Imediato |
| 2. Peticionamento | Preencher formulário online | 1-2 dias |
| 3. Pagamento | GRU (Guia de Recolhimento) | Imediato |
| 4. Depósito | Upload do código-fonte (hash ou resumo) | Imediato |
| 5. Certificado | Emissão do certificado | ~7 dias |

**Custos (2025)**

| Item | Valor (R$) | Desconto* |
|------|------------|-----------|
| Taxa de registro | R$ 185,00 | R$ 74,00 |
| Averbação de cessão | R$ 230,00 | R$ 92,00 |

*Desconto de 60% para MEI, ME, EPP, pessoas físicas, instituições de ensino e pesquisa, entidades sem fins lucrativos.

**Documentos Necessários**
- Descrição do software
- Código-fonte (até 720 KB) ou resumo digital (hash)
- Campos de aplicação
- Linguagem de programação
- Data de criação

### 2.3 Depósito do Código-Fonte

**Opções de depósito:**

1. **Código-fonte integral** (até 720 KB)
   - Vantagem: Prova completa
   - Desvantagem: Exposição do código

2. **Resumo digital (hash)**
   - Vantagem: Confidencialidade total
   - Desvantagem: Precisa guardar o código original

**Recomendação**: Usar hash SHA-256 do código-fonte compactado.

---

## 3. Conformidade LGPD

### 3.1 Visão Geral da LGPD

A **Lei Geral de Proteção de Dados (LGPD)** aplica-se a qualquer operação de tratamento de dados pessoais realizada no Brasil ou que ofereça serviços a indivíduos no Brasil.

### 3.2 Aplicabilidade ao Homelab

| Componente | Dados Pessoais? | Ação |
|------------|-----------------|------|
| Knative Lambda Operator | Possível (logs) | Anonimizar logs |
| Agent-Chat | Sim (mensagens) | Consentimento + criptografia |
| Agent-Medical | Sim (dados sensíveis) | Requisitos especiais |
| Agent-Restaurant | Sim (pedidos) | Política de privacidade |
| Agent-Contracts | Não (blockchain) | N/A |
| Observabilidade (Grafana) | Possível (IPs, logs) | Anonimização |

### 3.3 Bases Legais para Tratamento

| Base Legal | Quando Usar |
|------------|-------------|
| **Consentimento** | Funcionalidades opcionais, marketing |
| **Execução de Contrato** | Necessário para prestar o serviço |
| **Legítimo Interesse** | Segurança, prevenção de fraudes |
| **Obrigação Legal** | Cumprimento de leis |

### 3.4 Requisitos de Conformidade

#### Documentação Obrigatória

| Documento | Descrição |
|-----------|-----------|
| **Política de Privacidade** | Informações sobre coleta e uso de dados |
| **Termos de Uso** | Condições de uso do serviço |
| **RIPD** | Relatório de Impacto à Proteção de Dados |
| **Registro de Operações** | Documentação das atividades de tratamento |

#### Medidas Técnicas

- [ ] Criptografia de dados em trânsito (TLS)
- [ ] Criptografia de dados em repouso
- [ ] Controle de acesso (RBAC)
- [ ] Logs de auditoria
- [ ] Backup e recuperação
- [ ] Anonimização/pseudonimização

#### Direitos dos Titulares

Implementar mecanismos para:

| Direito | Implementação |
|---------|---------------|
| Acesso | API/Portal para visualizar dados |
| Correção | Funcionalidade de edição |
| Eliminação | Processo de exclusão |
| Portabilidade | Exportação em formato aberto |
| Revogação do consentimento | Opt-out fácil |

### 3.5 Penalidades LGPD

| Infração | Penalidade |
|----------|------------|
| Advertência | Prazo para correção |
| Multa simples | Até 2% do faturamento, limitado a R$ 50 milhões por infração |
| Multa diária | Valor definido pela ANPD |
| Publicização | Divulgação pública da infração |
| Bloqueio/Eliminação | Suspensão do tratamento |

### 3.6 Encarregado (DPO)

**Quando é obrigatório?**
- Tratamento em larga escala de dados sensíveis
- Monitoramento sistemático de titulares
- Atividade principal envolve tratamento de dados

**Recomendação**: Nomear DPO preventivamente e publicar contato no site.

---

## 4. Formação de Empresa

### 4.1 Tipos de Pessoa Jurídica

| Tipo | Características | Recomendado Para |
|------|-----------------|------------------|
| **MEI** | Faturamento até R$ 81k/ano, 1 pessoa | Início de operações |
| **LTDA** | 2+ sócios, responsabilidade limitada | Startups, PMEs |
| **EIRELI** | 1 sócio, capital mínimo 100 SM | Descontinuado (2021) |
| **SLU** | 1 sócio, sem capital mínimo | Empresário individual |
| **S.A.** | Estrutura complexa, ações | Grandes empresas, IPO |

### 4.2 Sociedade Limitada (LTDA) - Recomendada

**Vantagens:**
- Responsabilidade limitada ao capital social
- Flexibilidade na gestão
- Familiar para investidores
- Menor custo que S.A.

**Requisitos:**

| Requisito | Descrição |
|-----------|-----------|
| Sócios | Mínimo 2 (pode ser PF ou PJ) |
| Capital Social | Sem mínimo legal (prático: R$ 10.000+) |
| Contrato Social | Documento constitutivo |
| Sede | Endereço comercial no Brasil |
| CNPJ | Cadastro na Receita Federal |

**Custos de Abertura (estimativa)**

| Item | Valor |
|------|-------|
| Contador (abertura) | R$ 500 - 2.000 |
| Taxa Junta Comercial | R$ 200 - 500 |
| Certificado Digital | R$ 150 - 300 |
| Alvará de Funcionamento | R$ 100 - 500 |
| **Total Estimado** | **R$ 950 - 3.300** |

**Custos Mensais**

| Item | Valor |
|------|-------|
| Contador | R$ 300 - 1.500 |
| Impostos | Variável (ver seção 7) |

### 4.3 Processo de Abertura

1. **Consulta de viabilidade** - Verificar nome e endereço
2. **Elaborar Contrato Social** - Com advogado ou contador
3. **Registro na Junta Comercial** - NIRE
4. **Inscrição no CNPJ** - Receita Federal
5. **Inscrição Estadual** (se aplicável) - SEFAZ
6. **Inscrição Municipal** - Prefeitura
7. **Alvará de Funcionamento** - Prefeitura
8. **Certificado Digital** - Para emissão de NF-e

**Prazo**: 15-60 dias úteis

### 4.4 Sócio Estrangeiro

Se houver sócio estrangeiro:

- [ ] CPF para pessoa física estrangeira
- [ ] Procurador residente no Brasil
- [ ] Capital registrado no Banco Central (SISBACEN)
- [ ] Documentos traduzidos e notarizados

---

## 5. Propriedade Intelectual

### 5.1 Registro de Marca

**Por que registrar?**
- Uso exclusivo da marca no Brasil
- Proteção contra concorrentes
- Valorização do negócio
- Possibilidade de licenciamento

**Processo no INPI**

| Etapa | Descrição | Prazo |
|-------|-----------|-------|
| 1. Busca prévia | Verificar disponibilidade | 1-2 dias |
| 2. Pedido | Protocolar via e-INPI | Imediato |
| 3. Exame formal | Verificação de documentos | 1-3 meses |
| 4. Publicação | Revista da Propriedade Industrial | - |
| 5. Oposição | Terceiros podem se opor | 60 dias |
| 6. Exame de mérito | Análise técnica | 12-24 meses |
| 7. Deferimento | Aprovação | - |
| 8. Registro | Pagamento final + certificado | 60 dias |

**Custos (2025)**

| Item | Valor (R$) | Com Desconto* |
|------|------------|---------------|
| Pedido de registro | R$ 880/classe | R$ 440 |
| Expedição de certificado | Incluído | Incluído |
| Renovação (10 anos) | R$ 1.500/classe | R$ 750 |

*Desconto de 50% para pessoas físicas, MEI, ME, EPP, instituições de ensino/pesquisa.

**Classes Relevantes para Software**

| Classe | Descrição |
|--------|-----------|
| **Classe 9** | Software baixável, apps |
| **Classe 42** | SaaS, serviços de TI |
| **Classe 35** | Serviços de publicidade e gestão |

### 5.2 Busca de Anterioridade

Antes de registrar, pesquisar em:

1. **INPI**: [busca.inpi.gov.br](https://busca.inpi.gov.br/)
2. **Google**: Nomes similares
3. **Registro.br**: Domínios disponíveis
4. **Redes sociais**: Handles disponíveis

### 5.3 Patentes de Software

**No Brasil, software per se NÃO é patenteável** (Art. 10, Lei 9.279/1996).

**Pode ser patenteável**: Invenção implementada por computador que resolva problema técnico de forma nova e não óbvia.

**Alternativa**: Proteger como segredo industrial ou direito autoral.

---

## 6. Licenças Open Source

### 6.1 Validade no Brasil

Licenças open source são **contratos válidos** no Brasil, regidos pelo Código Civil.

| Licença | Tipo | Validade |
|---------|------|----------|
| MIT | Permissiva | ✅ Válida |
| Apache 2.0 | Permissiva | ✅ Válida |
| GPL v3 | Copyleft | ✅ Válida |
| AGPL v3 | Copyleft (SaaS) | ✅ Válida |

### 6.2 AGPL e SaaS no Brasil

A AGPL tem **validade e exigibilidade** no Brasil. Se você:

1. **Modifica software AGPL** e
2. **Oferece como serviço pela internet**

**Então**: Deve disponibilizar código-fonte modificado.

### 6.3 Situação do Slither-Analyzer

| Problema | slither-analyzer é AGPLv3 |
|----------|---------------------------|
| **Impacto** | Agent-Contracts como SaaS exige disclosure |
| **Solução A** | Licença comercial da Trail of Bits |
| **Solução B** | Substituir por Mythril (MIT) |
| **Solução C** | Manter Agent-Contracts open source |

---

## 7. Tributação

### 7.1 Regimes Tributários

| Regime | Faturamento Anual | Alíquota Efetiva |
|--------|-------------------|------------------|
| **Simples Nacional** | Até R$ 4,8 milhões | 6% - 33% |
| **Lucro Presumido** | Até R$ 78 milhões | ~15-25% |
| **Lucro Real** | Qualquer | ~34% + variáveis |

### 7.2 Simples Nacional para Software (Anexo V)

| Faixa | Receita Bruta (12 meses) | Alíquota |
|-------|--------------------------|----------|
| 1ª | Até R$ 180.000 | 15,50% |
| 2ª | R$ 180k - 360k | 18,00% |
| 3ª | R$ 360k - 720k | 19,50% |
| 4ª | R$ 720k - 1,8M | 20,50% |
| 5ª | R$ 1,8M - 3,6M | 23,00% |
| 6ª | R$ 3,6M - 4,8M | 30,50% |

**Nota**: Software como serviço (SaaS) geralmente enquadrado no Anexo III (mais favorável).

### 7.3 Impostos sobre SaaS

| Imposto | Alíquota | Observação |
|---------|----------|------------|
| **ISS** | 2% - 5% | Municipal, sobre serviços |
| **PIS** | 0,65% - 1,65% | Federal |
| **COFINS** | 3% - 7,6% | Federal |
| **IRPJ** | 15% + 10% | Lucro Real/Presumido |
| **CSLL** | 9% | Lucro Real/Presumido |

### 7.4 Venda de Software (Download)

| Situação | Tributação |
|----------|------------|
| Software "de prateleira" | ICMS (~18%) |
| Software customizado | ISS (2-5%) |
| SaaS | ISS (2-5%) |

---

## 8. Checklist de Conformidade

### 8.1 Pré-Lançamento

#### Proteção de PI

- [ ] Registrar software no INPI
- [ ] Protocolar pedido de marca no INPI
- [ ] Adicionar avisos de copyright ao código
- [ ] Criar arquivo NOTICE com atribuições

#### Conformidade LGPD

- [ ] Elaborar Política de Privacidade
- [ ] Elaborar Termos de Uso
- [ ] Implementar mecanismo de consentimento
- [ ] Criar processo para direitos dos titulares
- [ ] Nomear DPO (se aplicável)
- [ ] Realizar RIPD (se aplicável)

#### Licenças Open Source

- [ ] Resolver questão do Slither/AGPL
- [ ] Documentar todas as licenças de dependências
- [ ] Verificar compatibilidade de licenças

### 8.2 Estrutura Empresarial

- [ ] Definir tipo societário (LTDA recomendado)
- [ ] Elaborar Contrato Social
- [ ] Obter CNPJ
- [ ] Obter Inscrição Municipal
- [ ] Obter Alvará de Funcionamento
- [ ] Obter Certificado Digital
- [ ] Abrir conta bancária PJ

### 8.3 Operacional

- [ ] Contratar contador
- [ ] Definir regime tributário
- [ ] Implementar emissão de NFS-e
- [ ] Criar contratos de licenciamento
- [ ] Criar contratos de prestação de serviços

---

## Recursos Úteis

### Sites Oficiais

| Recurso | URL |
|---------|-----|
| INPI - Marcas | [inpi.gov.br/marcas](https://www.gov.br/inpi/pt-br/servicos/marcas) |
| INPI - Software | [inpi.gov.br/software](https://www.gov.br/inpi/pt-br/servicos/programas-de-computador) |
| ANPD | [gov.br/anpd](https://www.gov.br/anpd) |
| Receita Federal | [gov.br/receitafederal](https://www.gov.br/receitafederal) |
| Simples Nacional | [www8.receita.fazenda.gov.br/simplesnacional](http://www8.receita.fazenda.gov.br/simplesnacional/) |

### Legislação

| Lei | Link |
|-----|------|
| Lei do Software | [planalto.gov.br](http://www.planalto.gov.br/ccivil_03/leis/l9609.htm) |
| LGPD | [planalto.gov.br](http://www.planalto.gov.br/ccivil_03/_ato2015-2018/2018/lei/l13709.htm) |
| Lei de Propriedade Industrial | [planalto.gov.br](http://www.planalto.gov.br/ccivil_03/leis/l9279.htm) |

---

## Custos Estimados - Resumo

### Ano 1

| Item | Valor (R$) |
|------|------------|
| Abertura de empresa | 1.000 - 3.000 |
| Registro de software | 185 - 500 |
| Registro de marca (2 classes) | 1.760 - 3.000 |
| Contador (12 meses) | 3.600 - 18.000 |
| Consultoria jurídica | 2.000 - 10.000 |
| **Total Estimado** | **R$ 8.545 - 34.500** |

### Conversão Aproximada (USD)

| Item | USD (taxa ~5.0) |
|------|-----------------|
| Total Mínimo | ~$1.700 |
| Total Máximo | ~$6.900 |

---

**Documento preparado para**: Bruno Lucena / Projeto Homelab  
**Aviso Legal**: Este documento é apenas informativo e não constitui assessoria jurídica. Consulte um advogado especializado antes de tomar decisões comerciais.
