#!/bin/bash
# Copyright 2025 AutoPDF BuddhiLW
# SPDX-License-Identifier: Apache-2.0

# AutoPDF CLI Showcase - Demonstrates all capabilities of the SOLID + DDD + GoF refactored system

set -e

echo "🏗️  AutoPDF CLI Showcase - SOLID + DDD + GoF Architecture"
echo "================================================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Build the binary
echo -e "${BLUE}📦 Building AutoPDF...${NC}"
go build -o autopdf ./cmd/autopdf/main.go
echo -e "${GREEN}✅ Build complete${NC}"
echo ""

# Show help
echo -e "${BLUE}1️⃣  Showing help (SOLID + DDD + GoF documentation)${NC}"
./autopdf help
echo ""
echo -e "${GREEN}✅ Help shows architecture documentation${NC}"
echo ""

# Test with simple sample template
echo -e "${BLUE}2️⃣  Building simple sample template${NC}"
echo "   Template: templates/sample-template.tex"
echo "   Config: configs/sample-config.yaml"
echo ""
./autopdf build templates/sample-template.tex configs/sample-config.yaml
echo ""
echo -e "${GREEN}✅ Simple template built successfully${NC}"
echo ""

# Test with enhanced complex template
echo -e "${BLUE}3️⃣  Building enhanced complex template (nested data, loops, arrays)${NC}"
echo "   Template: templates/enhanced-document.tex"
echo "   Config: configs/enhanced-sample-config.yaml"
echo ""
./autopdf build templates/enhanced-document.tex configs/enhanced-sample-config.yaml
echo ""
echo -e "${GREEN}✅ Enhanced template with complex data structures built successfully${NC}"
echo ""

# Test with test-data complex config
echo -e "${BLUE}4️⃣  Building test-data complex template${NC}"
echo "   Template: internal/autopdf/test-data/complex_template.tex"
echo "   Config: internal/autopdf/test-data/complex_config.yaml"
echo ""
cd internal/autopdf/test-data
../../../autopdf build complex_template.tex complex_config.yaml
cd ../../..
echo ""
echo -e "${GREEN}✅ Test-data complex template built successfully${NC}"
echo ""

# Test with model_letter
echo -e "${BLUE}5️⃣  Building model letter (real-world example)${NC}"
echo "   Template: test/model_letter/main.tex"
echo "   Config: test/model_letter/config.yaml"
echo ""
cd test/model_letter
../../autopdf build main.tex config.yaml
cd ../..
echo ""
echo -e "${GREEN}✅ Model letter built successfully${NC}"
echo ""

# Test with model_xelatex
echo -e "${BLUE}6️⃣  Building XeLaTeX model (demonstrates engine selection)${NC}"
echo "   Template: test/model_xelatex/main.tex"
echo "   Config: test/model_xelatex/config.yaml"
echo "   Engine: xelatex (Factory Pattern in action)"
echo ""
cd test/model_xelatex
../../autopdf build main.tex config.yaml
cd ../..
echo ""
echo -e "${GREEN}✅ XeLaTeX model built successfully (Factory Pattern)${NC}"
echo ""

# Test clean command
echo -e "${BLUE}7️⃣  Testing clean command (Domain Layer)${NC}"
echo "   Cleaning auxiliary files..."
echo ""
./autopdf clean internal/autopdf/test-data/out
echo ""
echo -e "${GREEN}✅ Clean command executed (FileManagementService)${NC}"
echo ""

# Test conversion command (if ImageMagick/Poppler available)
echo -e "${BLUE}8️⃣  Testing conversion command (Strategy Pattern)${NC}"
echo "   Converting PDF to images..."
echo ""
if [ -f "output/final.pdf" ]; then
    ./autopdf convert output/final.pdf png || echo -e "${YELLOW}⚠️  Conversion tools not available (ImageMagick/Poppler needed)${NC}"
else
    echo -e "${YELLOW}⚠️  No PDF available for conversion test${NC}"
fi
echo ""

# Summary
echo ""
echo "================================================================"
echo -e "${GREEN}🎉 CLI Showcase Complete!${NC}"
echo ""
echo "Demonstrated Features:"
echo "  ✓ Simple templates with basic variables"
echo "  ✓ Complex templates with nested data structures"
echo "  ✓ Arrays and loops in templates"
echo "  ✓ Multiple LaTeX engines (pdflatex, xelatex)"
echo "  ✓ Real-world examples (letters, documents)"
echo "  ✓ Clean command (Domain Services)"
echo "  ✓ Conversion command (Strategy Pattern)"
echo ""
echo "Architecture Highlights:"
echo "  • SOLID Principles: Single Responsibility, Dependency Inversion"
echo "  • DDD: Domain Services, Value Objects, Entities"
echo "  • GoF Patterns: Factory, Strategy, Observer, Facade"
echo "  • Event-Driven: Observable build process"
echo "  • Testable: All services mockable"
echo ""
echo "Generated Files:"
find output -name "*.pdf" 2>/dev/null | head -5 || echo "  (Check output directories for generated PDFs)"
echo ""
echo "To see detailed architecture documentation:"
echo "  ./autopdf help"
echo ""
echo "To run integration tests:"
echo "  go test ./test/integration/... -v"
echo ""
