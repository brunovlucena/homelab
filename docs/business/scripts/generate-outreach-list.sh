#!/bin/bash
# Generate Outreach List from Target Customers
# Creates CSV file with all target customers for CRM import

set -e

OUTPUT_FILE="target_customers.csv"

echo "Generating target customer list for CRM import..."

cat > "$OUTPUT_FILE" << 'EOF'
Company,Market,Industry,Address,Phone,Priority,Decision Maker,Email,LinkedIn,Notes,Next Action,Date
Real Hospital Português,Recife,Healthcare,Av. Agamenon Magalhães 4760 - Paissandu,+55 81 3416-1000,HIGH,IT Director,research@hospital.com,research LinkedIn,,"Send email","2025-01-XX"
Hospital Memorial São José,Recife,Healthcare,Av. Gov. Agamenon Magalhães 2291 - Derby,+55 81 3216-2222,HIGH,IT Director,research@hospital.com,research LinkedIn,,"Send email","2025-01-XX"
Hospital Esperança Recife,Recife,Healthcare,R. Antônio Gomes de Freitas 265 - Ilha do Leite,+55 81 3421-5000,MEDIUM,IT Director,research@hospital.com,research LinkedIn,,"Send email","2025-01-XX"
Massachusetts General Hospital,Boston,Healthcare,55 Fruit St Boston MA 02114,617-726-2000,HIGH,IT/Innovation Director,research@hospital.com,research LinkedIn,,"Send email","2025-01-XX"
Brigham and Women's Hospital,Boston,Healthcare,75 Francis St Boston MA 02115,617-732-5500,HIGH,IT/Innovation Director,research@hospital.com,research LinkedIn,,"Send email","2025-01-XX"
Beth Israel Deaconess Medical Center,Boston,Healthcare,330 Brookline Ave Boston MA 02215,617-667-7000,HIGH,IT/Innovation Director,research@hospital.com,research LinkedIn,,"Send email","2025-01-XX"
McDonald's Pina,Recife,Restaurant,Av. República do Líbano 251 - Pina,(81) 3033-2335,HIGH,Operations Manager,research@mcdonalds.com,research LinkedIn,,"Send email","2025-01-XX"
McDonald's Boa Viagem,Recife,Restaurant,R. Dom João VI 570 - Boa Viagem,(81) 3326-3902,HIGH,Operations Manager,research@mcdonalds.com,research LinkedIn,,"Send email","2025-01-XX"
McDonald's Commonwealth Ave,Boston,Restaurant,540 Commonwealth Ave Boston MA 02215,-,HIGH,Operations Manager,research@mcdonalds.com,research LinkedIn,,"Send email","2025-01-XX"
McDonald's Tremont Street,Boston,Restaurant,178 Tremont Street Boston MA 02111,-,HIGH,Operations Manager,research@mcdonalds.com,research LinkedIn,,"Send email","2025-01-XX"
Petro Mix Chain,Recife,Gas Station,Multiple locations,-,HIGH,Operations Manager,research@petromix.com,research LinkedIn,,"Send email","2025-01-XX"
Shell Station,Boston,Gas Station,1001 Massachusetts Ave Boston MA 02118,-,HIGH,Operations Manager,research@shell.com,research LinkedIn,,"Send email","2025-01-XX"
Mobil Station,Boston,Gas Station,273 E Berkeley St Boston MA 02118,-,HIGH,Operations Manager,research@mobil.com,research LinkedIn,,"Send email","2025-01-XX"
EOF

echo "✅ Generated $OUTPUT_FILE"
echo ""
echo "Next steps:"
echo "1. Open $OUTPUT_FILE in Excel/Google Sheets"
echo "2. Research and fill in: Decision Maker, Email, LinkedIn"
echo "3. Import to CRM (HubSpot, Pipedrive, etc.)"
echo ""
