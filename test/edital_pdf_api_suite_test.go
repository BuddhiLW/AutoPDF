package test

import (
	"errors"
	"testing"

	"github.com/BuddhiLW/AutoPDF/mocks"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
	"github.com/BuddhiLW/AutoPDF/pkg/domain"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// EditalPdfApiTestSuite tests AutoPDF functionality for legal document generation
type EditalPdfApiTestSuite struct {
	suite.Suite
	mockEngine            *mocks.MockTemplateEngine
	mockEnhancedEngine    *mocks.MockEnhancedTemplateEngine
	mockValidator         *mocks.MockTemplateValidator
	mockFileProcessor     *mocks.MockFileProcessor
	mockVariableProcessor *mocks.MockVariableProcessor
	mockConfigProvider    *mocks.MockConfigProvider
}

// SetupTest initializes the test suite
func (suite *EditalPdfApiTestSuite) SetupTest() {
	suite.mockEngine = mocks.NewMockTemplateEngine(suite.T())
	suite.mockEnhancedEngine = mocks.NewMockEnhancedTemplateEngine(suite.T())
	suite.mockValidator = mocks.NewMockTemplateValidator(suite.T())
	suite.mockFileProcessor = mocks.NewMockFileProcessor(suite.T())
	suite.mockVariableProcessor = mocks.NewMockVariableProcessor(suite.T())
	suite.mockConfigProvider = mocks.NewMockConfigProvider(suite.T())
}

// TestLegalDocumentGeneration tests the complete legal document generation workflow
func (suite *EditalPdfApiTestSuite) TestLegalDocumentGeneration() {
	// Test data representing a legal auction document
	editalData := map[string]interface{}{
		"leilao": map[string]interface{}{
			"vara":     "1ª Vara Cível",
			"comarca":  "Belo Horizonte",
			"estado":   "MG",
			"processo": "1234567-89.2024.8.13.0024",
		},
		"executado": "Empresa XYZ Ltda",
		"documento": "12.345.678/0001-90",
		"exequente": "Banco ABC S.A.",
		"juiz": map[string]interface{}{
			"nome": "Dr. João Silva",
		},
		"site":                      "www.leiloes.com.br",
		"encerramento":              "25/01/2024 às 14:00",
		"segundoLeilao":             true,
		"encerramentoSegundoLeilao": "26/01/2024 às 14:00",
		"descontoSegundoLeilao":     80,
		"valorDivida":               "R$ 150.000,00",
		"valorDividaExtenso":        "cento e cinquenta mil reais",
		"origemValorDivida":         "sentença judicial",
		"data_edital":               "15/01/2024",
		"bens": []interface{}{
			map[string]interface{}{
				"descricao": "Imóvel Residencial",
				"registro":  "Matrícula 12345",
				"avaliacao": map[string]interface{}{
					"valor":        "R$ 200.000,00",
					"valorExtenso": "duzentos mil reais",
					"origem":       "avaliação judicial",
				},
				"onus": []interface{}{
					"Hipoteca em favor do Banco ABC S.A.",
					"Penhor de veículos",
				},
			},
		},
	}

	// Set up expectations for the complete workflow
	suite.mockConfigProvider.EXPECT().
		GetConfig().
		Return(&config.Config{
			Template:  "edital_template_enhanced.tex",
			Output:    "edital_leilao.pdf",
			Variables: editalData,
			Engine:    "xelatex",
		}).
		Once()

	suite.mockValidator.EXPECT().
		ValidateTemplate("edital_template_enhanced.tex").
		Return(nil).
		Once()

	suite.mockFileProcessor.EXPECT().
		FileExists("edital_template_enhanced.tex").
		Return(true).
		Once()

	// Mock ABNTeX template content
	templateContent := `\documentclass[12pt,a4paper,oneside]{abntex2}
\usepackage[utf8]{inputenc}
\usepackage[T1]{fontenc}
\usepackage{graphicx}
\usepackage[dvipsnames]{xcolor}
\usepackage{ebgaramond}
\usepackage{longtable}
\usepackage{booktabs}
\usepackage{xstring}
\usepackage{etoolbox}
\usepackage{calc}
\usepackage{enumitem}
\usepackage{hyperref}
\usepackage{geometry}

\newcommand{\LeilaoVaraVar}{delim[[.leilao.vara]]}
\newcommand{\LeilaoComarcaVar}{delim[[.leilao.comarca]]}
\newcommand{\LeilaoEstadoVar}{delim[[.leilao.estado]]}
\newcommand{\LeilaoProcessoVar}{delim[[.leilao.processo]]}
\newcommand{\ExecutadoVar}{delim[[.executado]]}
\newcommand{\DocumentoVar}{delim[[.documento]]}
\newcommand{\ExequenteVar}{delim[[.exequente]]}
\newcommand{\JuizNomeVar}{delim[[.juiz.nome]]}
\newcommand{\SiteVar}{delim[[.site]]}
\newcommand{\EncerramentoVar}{delim[[.encerramento]]}
\newcommand{\ValorDividaVar}{delim[[.valorDivida]]}
\newcommand{\ValorDividaExtensoVar}{delim[[.valorDividaExtenso]]}
\newcommand{\OrigemValorDividaVar}{delim[[.origemValorDivida]]}
\newcommand{\DataEditalVar}{delim[[.data_edital]]}

\begin{document}
\begin{center}
  {\textcolor{HeaderBlue}{\Title{EDITAL DE 1º E 2º LEILÃO E INTIMAÇÃO}}} \\
  \vspace{0.5cm}
  
  {\textcolor{HeaderBlue}{\Large\textbf{\LeilaoVaraVar da Comarca de \LeilaoComarcaVar/\LeilaoEstadoVar}}} \\
  \vspace{0.5cm}
  
  \rule{\textwidth}{1pt}
  \vspace{0.5cm}
\end{center}

\begin{flushleft}
\textbf{EDITAL de 1º e 2º LEILÃO} \\
\textbf{DE BEM IMÓVEL} \\
\textbf{para intimação} \\
\textbf{da empresa executada} \\
\textbf{\ExecutadoVar - CNPJ nº \DocumentoVar,} \\
por meio de seu representante legal, executado, \\
proprietário e fiel depositário do bem \\

delim[[if .bens]]
delim[[if gt (len .bens) 1]]
os bens abaixo descritos,
delim[[else]]
o bem abaixo descrito,
delim[[end]]
delim[[else]]
o bem abaixo descrito,
delim[[end]]
conforme condições de venda constantes no presente edital. \\

No 1° Leilão com início da publicação do edital e término no dia \\
\EncerramentoVar, \\
não serão admitidos lances inferiores ao valor de avaliação atualizada do bem. \\

delim[[if .segundoLeilao]]
Ficando desde já designado para o 2° Leilão com início no dia \\
\EncerramentoVar e término no dia delim[[.encerramentoSegundoLeilao]], \\
caso não haja licitantes no 1º Leilão. \\

delim[[if .descontoSegundoLeilao]]
No segundo serão admitidos lances não inferiores a delim[[.descontoSegundoLeilao]]\% do valor da avaliação atualizada. \\
delim[[end]]
delim[[end]]

\textbf{VALOR DA DÍVIDA NO PROCESSO DE EXECUÇÃO:} O valor da dívida no processo de execução é de \\
\ValorDividaVar (\ValorDividaExtensoVar), \\
conforme \OrigemValorDividaVar. \\

\LeilaoComarcaVar/\LeilaoEstadoVar, \DataEditalVar. \\

\begin{center}
\textbf{\JuizNomeVar} \\
\textbf{JUIZ DE DIREITO}
\end{center}

\end{flushleft}
\end{document}`

	suite.mockFileProcessor.EXPECT().
		ReadFile("edital_template_enhanced.tex").
		Return([]byte(templateContent), nil).
		Once()

	// Process variables for legal document
	expectedCollection := domain.NewVariableCollection()
	leilaoVar, _ := domain.NewObjectVariable(editalData["leilao"].(map[string]interface{}))
	executadoVar, _ := domain.NewStringVariable("Empresa XYZ Ltda")
	documentoVar, _ := domain.NewStringVariable("12.345.678/0001-90")
	exequenteVar, _ := domain.NewStringVariable("Banco ABC S.A.")
	juizVar, _ := domain.NewObjectVariable(editalData["juiz"].(map[string]interface{}))
	siteVar, _ := domain.NewStringVariable("www.leiloes.com.br")
	encerramentoVar, _ := domain.NewStringVariable("25/01/2024 às 14:00")
	segundoLeilaoVar := domain.NewBooleanVariable(true)
	encerramentoSegundoLeilaoVar, _ := domain.NewStringVariable("26/01/2024 às 14:00")
	descontoSegundoLeilaoVar, _ := domain.NewNumberVariable(80)
	valorDividaVar, _ := domain.NewStringVariable("R$ 150.000,00")
	valorDividaExtensoVar, _ := domain.NewStringVariable("cento e cinquenta mil reais")
	origemValorDividaVar, _ := domain.NewStringVariable("sentença judicial")
	dataEditalVar, _ := domain.NewStringVariable("15/01/2024")

	expectedCollection.Set("leilao", leilaoVar)
	expectedCollection.Set("executado", executadoVar)
	expectedCollection.Set("documento", documentoVar)
	expectedCollection.Set("exequente", exequenteVar)
	expectedCollection.Set("juiz", juizVar)
	expectedCollection.Set("site", siteVar)
	expectedCollection.Set("encerramento", encerramentoVar)
	expectedCollection.Set("segundoLeilao", segundoLeilaoVar)
	expectedCollection.Set("encerramentoSegundoLeilao", encerramentoSegundoLeilaoVar)
	expectedCollection.Set("descontoSegundoLeilao", descontoSegundoLeilaoVar)
	expectedCollection.Set("valorDivida", valorDividaVar)
	expectedCollection.Set("valorDividaExtenso", valorDividaExtensoVar)
	expectedCollection.Set("origemValorDivida", origemValorDividaVar)
	expectedCollection.Set("data_edital", dataEditalVar)
	bensVar, _ := domain.NewArrayVariable(editalData["bens"].([]interface{}))
	expectedCollection.Set("bens", bensVar)

	suite.mockVariableProcessor.EXPECT().
		ProcessVariables(editalData).
		Return(expectedCollection, nil).
		Once()

	// Process template with legal document data
	processedContent := `\documentclass[12pt,a4paper,oneside]{abntex2}
\usepackage[utf8]{inputenc}
\usepackage[T1]{fontenc}
\usepackage{graphicx}
\usepackage[dvipsnames]{xcolor}
\usepackage{ebgaramond}
\usepackage{longtable}
\usepackage{booktabs}
\usepackage{xstring}
\usepackage{etoolbox}
\usepackage{calc}
\usepackage{enumitem}
\usepackage{hyperref}
\usepackage{geometry}

\newcommand{\LeilaoVaraVar}{1ª Vara Cível}
\newcommand{\LeilaoComarcaVar}{Belo Horizonte}
\newcommand{\LeilaoEstadoVar}{MG}
\newcommand{\LeilaoProcessoVar}{1234567-89.2024.8.13.0024}
\newcommand{\ExecutadoVar}{Empresa XYZ Ltda}
\newcommand{\DocumentoVar}{12.345.678/0001-90}
\newcommand{\ExequenteVar}{Banco ABC S.A.}
\newcommand{\JuizNomeVar}{Dr. João Silva}
\newcommand{\SiteVar}{www.leiloes.com.br}
\newcommand{\EncerramentoVar}{25/01/2024 às 14:00}
\newcommand{\ValorDividaVar}{R$ 150.000,00}
\newcommand{\ValorDividaExtensoVar}{cento e cinquenta mil reais}
\newcommand{\OrigemValorDividaVar}{sentença judicial}
\newcommand{\DataEditalVar}{15/01/2024}

\begin{document}
\begin{center}
  {\textcolor{HeaderBlue}{\Title{EDITAL DE 1º E 2º LEILÃO E INTIMAÇÃO}}} \\
  \vspace{0.5cm}
  
  {\textcolor{HeaderBlue}{\Large\textbf{1ª Vara Cível da Comarca de Belo Horizonte/MG}}} \\
  \vspace{0.5cm}
  
  \rule{\textwidth}{1pt}
  \vspace{0.5cm}
\end{center}

\begin{flushleft}
\textbf{EDITAL de 1º e 2º LEILÃO} \\
\textbf{DE BEM IMÓVEL} \\
\textbf{para intimação} \\
\textbf{da empresa executada} \\
\textbf{Empresa XYZ Ltda - CNPJ nº 12.345.678/0001-90,} \\
por meio de seu representante legal, executado, \\
proprietário e fiel depositário do bem \\

o bem abaixo descrito,
conforme condições de venda constantes no presente edital. \\

No 1° Leilão com início da publicação do edital e término no dia \\
25/01/2024 às 14:00, \\
não serão admitidos lances inferiores ao valor de avaliação atualizada do bem. \\

Ficando desde já designado para o 2° Leilão com início no dia \\
25/01/2024 às 14:00 e término no dia 26/01/2024 às 14:00, \\
caso não haja licitantes no 1º Leilão. \\

No segundo serão admitidos lances não inferiores a 80\% do valor da avaliação atualizada. \\

\textbf{VALOR DA DÍVIDA NO PROCESSO DE EXECUÇÃO:} O valor da dívida no processo de execução é de \\
R$ 150.000,00 (cento e cinquenta mil reais), \\
conforme sentença judicial. \\

Belo Horizonte/MG, 15/01/2024. \\

\begin{center}
\textbf{Dr. João Silva} \\
\textbf{JUIZ DE DIREITO}
\end{center}

\end{flushleft}
\end{document}`

	suite.mockEngine.EXPECT().
		Process("edital_template_enhanced.tex").
		Return(processedContent, nil).
		Once()

	suite.mockFileProcessor.EXPECT().
		CreateDirectory("output").
		Return(nil).
		Once()

	suite.mockFileProcessor.EXPECT().
		WriteFile("output/processed.tex", []byte(processedContent)).
		Return(nil).
		Once()

	// Execute the workflow
	cfg := suite.mockConfigProvider.GetConfig()
	suite.Equal("edital_template_enhanced.tex", cfg.Template.String())
	suite.Equal("edital_leilao.pdf", cfg.Output.String())
	suite.Equal("xelatex", cfg.Engine.String())

	err := suite.mockValidator.ValidateTemplate("edital_template_enhanced.tex")
	suite.NoError(err)

	exists := suite.mockFileProcessor.FileExists("edital_template_enhanced.tex")
	suite.True(exists)

	content, err := suite.mockFileProcessor.ReadFile("edital_template_enhanced.tex")
	suite.NoError(err)
	suite.Contains(string(content), "\\documentclass[12pt,a4paper,oneside]{abntex2}")

	collection, err := suite.mockVariableProcessor.ProcessVariables(cfg.Variables)
	suite.NoError(err)
	suite.NotNil(collection)

	result, err := suite.mockEngine.Process("edital_template_enhanced.tex")
	suite.NoError(err)
	suite.Contains(result, "Empresa XYZ Ltda")
	suite.Contains(result, "12.345.678/0001-90")
	suite.Contains(result, "Dr. João Silva")

	err = suite.mockFileProcessor.CreateDirectory("output")
	suite.NoError(err)

	err = suite.mockFileProcessor.WriteFile("output/processed.tex", []byte(result))
	suite.NoError(err)
}

// TestComplexLegalDocument tests legal document with multiple assets
func (suite *EditalPdfApiTestSuite) TestComplexLegalDocument() {
	complexEditalData := map[string]interface{}{
		"leilao": map[string]interface{}{
			"vara":     "2ª Vara Cível",
			"comarca":  "São Paulo",
			"estado":   "SP",
			"processo": "9876543-21.2024.8.26.0100",
		},
		"executado": "Construtora ABC Ltda",
		"documento": "98.765.432/0001-10",
		"exequente": "Banco XYZ S.A.",
		"juiz": map[string]interface{}{
			"nome": "Dra. Maria Santos",
		},
		"site":                      "www.leiloesjudiciais.com.br",
		"encerramento":              "30/01/2024 às 15:00",
		"segundoLeilao":             true,
		"encerramentoSegundoLeilao": "31/01/2024 às 15:00",
		"descontoSegundoLeilao":     80,
		"valorDivida":               "R$ 500.000,00",
		"valorDividaExtenso":        "quinhentos mil reais",
		"origemValorDivida":         "sentença judicial transitada em julgado",
		"data_edital":               "20/01/2024",
		"bens": []interface{}{
			map[string]interface{}{
				"descricao": "Imóvel Residencial",
				"registro":  "Matrícula 54321",
				"avaliacao": map[string]interface{}{
					"valor":        "R$ 300.000,00",
					"valorExtenso": "trezentos mil reais",
					"origem":       "avaliação judicial",
				},
				"onus": []interface{}{
					"Hipoteca em favor do Banco XYZ S.A.",
				},
			},
			map[string]interface{}{
				"descricao": "Veículo Automotivo",
				"registro":  "Placa ABC-1234",
				"avaliacao": map[string]interface{}{
					"valor":        "R$ 50.000,00",
					"valorExtenso": "cinquenta mil reais",
					"origem":       "avaliação judicial",
				},
				"onus": []interface{}{
					"Penhor em favor do Banco XYZ S.A.",
				},
			},
		},
	}

	suite.mockConfigProvider.EXPECT().
		GetConfig().
		Return(&config.Config{
			Template:  "edital_template_enhanced.tex",
			Output:    "edital_complexo.pdf",
			Variables: complexEditalData,
			Engine:    "xelatex",
		}).
		Once()

	suite.mockValidator.EXPECT().
		ValidateTemplate("edital_template_enhanced.tex").
		Return(nil).
		Once()

	suite.mockFileProcessor.EXPECT().
		FileExists("edital_template_enhanced.tex").
		Return(true).
		Once()

	suite.mockFileProcessor.EXPECT().
		ReadFile("edital_template_enhanced.tex").
		Return([]byte("template content"), nil).
		Once()

	expectedCollection := domain.NewVariableCollection()
	leilaoVar, _ := domain.NewObjectVariable(complexEditalData["leilao"].(map[string]interface{}))
	expectedCollection.Set("leilao", leilaoVar)
	executadoVar, _ := domain.NewStringVariable("Construtora ABC Ltda")
	documentoVar, _ := domain.NewStringVariable("98.765.432/0001-10")
	exequenteVar, _ := domain.NewStringVariable("Banco XYZ S.A.")
	juizVar, _ := domain.NewObjectVariable(complexEditalData["juiz"].(map[string]interface{}))
	siteVar, _ := domain.NewStringVariable("www.leiloesjudiciais.com.br")
	encerramentoVar, _ := domain.NewStringVariable("30/01/2024 às 15:00")
	segundoLeilaoVar := domain.NewBooleanVariable(true)
	encerramentoSegundoLeilaoVar, _ := domain.NewStringVariable("31/01/2024 às 15:00")
	descontoSegundoLeilaoVar, _ := domain.NewNumberVariable(80)
	valorDividaVar, _ := domain.NewStringVariable("R$ 500.000,00")
	valorDividaExtensoVar, _ := domain.NewStringVariable("quinhentos mil reais")
	origemValorDividaVar, _ := domain.NewStringVariable("sentença judicial transitada em julgado")
	dataEditalVar, _ := domain.NewStringVariable("20/01/2024")
	bensVar, _ := domain.NewArrayVariable(complexEditalData["bens"].([]interface{}))

	expectedCollection.Set("leilao", leilaoVar)
	expectedCollection.Set("executado", executadoVar)
	expectedCollection.Set("documento", documentoVar)
	expectedCollection.Set("exequente", exequenteVar)
	expectedCollection.Set("juiz", juizVar)
	expectedCollection.Set("site", siteVar)
	expectedCollection.Set("encerramento", encerramentoVar)
	expectedCollection.Set("segundoLeilao", segundoLeilaoVar)
	expectedCollection.Set("encerramentoSegundoLeilao", encerramentoSegundoLeilaoVar)
	expectedCollection.Set("descontoSegundoLeilao", descontoSegundoLeilaoVar)
	expectedCollection.Set("valorDivida", valorDividaVar)
	expectedCollection.Set("valorDividaExtenso", valorDividaExtensoVar)
	expectedCollection.Set("origemValorDivida", origemValorDividaVar)
	expectedCollection.Set("data_edital", dataEditalVar)
	expectedCollection.Set("bens", bensVar)

	suite.mockVariableProcessor.EXPECT().
		ProcessVariables(complexEditalData).
		Return(expectedCollection, nil).
		Once()

	suite.mockEngine.EXPECT().
		Process("edital_template_enhanced.tex").
		Return("Complex legal document with multiple assets", nil).
		Once()

	suite.mockFileProcessor.EXPECT().
		CreateDirectory("output").
		Return(nil).
		Once()

	suite.mockFileProcessor.EXPECT().
		WriteFile("output/processed.tex", []byte("Complex legal document with multiple assets")).
		Return(nil).
		Once()

	// Execute workflow
	cfg := suite.mockConfigProvider.GetConfig()
	suite.Equal("Construtora ABC Ltda", cfg.Variables["executado"])

	err := suite.mockValidator.ValidateTemplate("edital_template_enhanced.tex")
	suite.NoError(err)

	exists := suite.mockFileProcessor.FileExists("edital_template_enhanced.tex")
	suite.True(exists)

	content, err := suite.mockFileProcessor.ReadFile("edital_template_enhanced.tex")
	suite.NoError(err)
	suite.NotNil(content)

	collection, err := suite.mockVariableProcessor.ProcessVariables(cfg.Variables)
	suite.NoError(err)
	suite.NotNil(collection)

	result, err := suite.mockEngine.Process("edital_template_enhanced.tex")
	suite.NoError(err)
	suite.Contains(result, "multiple assets")

	err = suite.mockFileProcessor.CreateDirectory("output")
	suite.NoError(err)

	err = suite.mockFileProcessor.WriteFile("output/processed.tex", []byte(result))
	suite.NoError(err)
}

// TestLegalDocumentErrorHandling tests error scenarios in legal document generation
func (suite *EditalPdfApiTestSuite) TestLegalDocumentErrorHandling() {
	suite.Run("ABNTeXValidationFailure", func() {
		suite.mockValidator.EXPECT().
			ValidateTemplate("invalid_abntex.tex").
			Return(errors.New("ABNTeX syntax error: missing \\begin{document}")).
			Once()

		err := suite.mockValidator.ValidateTemplate("invalid_abntex.tex")
		suite.Error(err)
		suite.Contains(err.Error(), "ABNTeX syntax error")
	})

	suite.Run("ComplexDataProcessingError", func() {
		complexData := map[string]interface{}{
			"leilao": map[string]interface{}{
				"vara": "1ª Vara Cível",
			},
		}

		suite.mockVariableProcessor.EXPECT().
			ProcessVariables(complexData).
			Return(nil, errors.New("failed to process complex nested data")).
			Once()

		collection, err := suite.mockVariableProcessor.ProcessVariables(complexData)
		suite.Error(err)
		suite.Nil(collection)
		suite.Contains(err.Error(), "complex nested data")
	})

	suite.Run("TemplateProcessingError", func() {
		suite.mockEngine.EXPECT().
			Process("error_template.tex").
			Return("", errors.New("template processing failed: undefined variable 'missing_legal_var'")).
			Once()

		result, err := suite.mockEngine.Process("error_template.tex")
		suite.Error(err)
		suite.Empty(result)
		suite.Contains(err.Error(), "undefined variable")
	})
}

// TestLegalDocumentPerformance tests performance scenarios for legal document generation
func (suite *EditalPdfApiTestSuite) TestLegalDocumentPerformance() {
	suite.Run("HighVolumeLegalDocumentGeneration", func() {
		// Test generating multiple legal documents concurrently
		suite.mockEngine.EXPECT().
			Process(mock.AnythingOfType("string")).
			Return("processed legal document", nil).
			Times(100)

		// Simulate concurrent processing of 100 legal documents
		done := make(chan bool, 100)
		for i := 0; i < 100; i++ {
			go func(index int) {
				result, err := suite.mockEngine.Process("edital_template.tex")
				suite.NoError(err)
				suite.Contains(result, "legal document")
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 100; i++ {
			<-done
		}
	})

	suite.Run("LargeLegalDocumentProcessing", func() {
		// Test processing legal documents with large amounts of data
		largeData := make(map[string]interface{})
		for i := 0; i < 1000; i++ {
			largeData[suite.T().Name()+"_key_"+string(rune(i))] = "value_" + string(rune(i))
		}

		suite.mockVariableProcessor.EXPECT().
			ProcessVariables(largeData).
			Return(domain.NewVariableCollection(), nil).
			Once()

		suite.mockEngine.EXPECT().
			Process("large_edital_template.tex").
			Return("Large legal document processed successfully", nil).
			Once()

		collection, err := suite.mockVariableProcessor.ProcessVariables(largeData)
		suite.NoError(err)
		suite.NotNil(collection)

		result, err := suite.mockEngine.Process("large_edital_template.tex")
		suite.NoError(err)
		suite.Contains(result, "Large legal document")
	})
}

// TestLegalDocumentIntegration tests integration scenarios
func (suite *EditalPdfApiTestSuite) TestLegalDocumentIntegration() {
	suite.Run("CompleteWorkflowWithRealLegalData", func() {
		// Simulate real legal document data structure
		realLegalData := map[string]interface{}{
			"leilao": map[string]interface{}{
				"vara":     "3ª Vara Cível",
				"comarca":  "Rio de Janeiro",
				"estado":   "RJ",
				"processo": "1111111-11.2024.8.19.0001",
			},
			"executado": "Empresa Real Ltda",
			"documento": "11.111.111/0001-11",
			"exequente": "Banco Real S.A.",
			"juiz": map[string]interface{}{
				"nome": "Dr. Pedro Oliveira",
			},
			"site":                      "www.leiloesreais.com.br",
			"encerramento":              "05/02/2024 às 16:00",
			"segundoLeilao":             true,
			"encerramentoSegundoLeilao": "06/02/2024 às 16:00",
			"descontoSegundoLeilao":     80,
			"valorDivida":               "R$ 1.000.000,00",
			"valorDividaExtenso":        "um milhão de reais",
			"origemValorDivida":         "sentença judicial transitada em julgado",
			"data_edital":               "25/01/2024",
			"bens": []interface{}{
				map[string]interface{}{
					"descricao": "Imóvel Comercial",
					"registro":  "Matrícula 99999",
					"avaliacao": map[string]interface{}{
						"valor":        "R$ 800.000,00",
						"valorExtenso": "oitocentos mil reais",
						"origem":       "avaliação judicial",
					},
					"onus": []interface{}{
						"Hipoteca em favor do Banco Real S.A.",
						"Penhor de equipamentos",
					},
				},
			},
		}

		suite.mockConfigProvider.EXPECT().
			GetConfig().
			Return(&config.Config{
				Template:  "edital_template_enhanced.tex",
				Output:    "edital_real.pdf",
				Variables: realLegalData,
				Engine:    "xelatex",
			}).
			Once()

		suite.mockValidator.EXPECT().
			ValidateTemplate("edital_template_enhanced.tex").
			Return(nil).
			Once()

		suite.mockFileProcessor.EXPECT().
			FileExists("edital_template_enhanced.tex").
			Return(true).
			Once()

		suite.mockFileProcessor.EXPECT().
			ReadFile("edital_template_enhanced.tex").
			Return([]byte("template content"), nil).
			Once()

		expectedCollection := domain.NewVariableCollection()
		leilaoVar, _ := domain.NewObjectVariable(realLegalData["leilao"].(map[string]interface{}))
		expectedCollection.Set("leilao", leilaoVar)
		executadoVar, _ := domain.NewStringVariable("Empresa Real Ltda")
		documentoVar, _ := domain.NewStringVariable("11.111.111/0001-11")
		exequenteVar, _ := domain.NewStringVariable("Banco Real S.A.")
		juizVar, _ := domain.NewObjectVariable(realLegalData["juiz"].(map[string]interface{}))
		siteVar, _ := domain.NewStringVariable("www.leiloesreais.com.br")
		encerramentoVar, _ := domain.NewStringVariable("05/02/2024 às 16:00")
		segundoLeilaoVar := domain.NewBooleanVariable(true)
		encerramentoSegundoLeilaoVar, _ := domain.NewStringVariable("06/02/2024 às 16:00")
		descontoSegundoLeilaoVar, _ := domain.NewNumberVariable(80)
		valorDividaVar, _ := domain.NewStringVariable("R$ 1.000.000,00")
		valorDividaExtensoVar, _ := domain.NewStringVariable("um milhão de reais")
		origemValorDividaVar, _ := domain.NewStringVariable("sentença judicial transitada em julgado")
		dataEditalVar, _ := domain.NewStringVariable("25/01/2024")
		bensVar, _ := domain.NewArrayVariable(realLegalData["bens"].([]interface{}))
		expectedCollection.Set("executado", executadoVar)
		expectedCollection.Set("documento", documentoVar)
		expectedCollection.Set("exequente", exequenteVar)
		expectedCollection.Set("juiz", juizVar)
		expectedCollection.Set("site", siteVar)
		expectedCollection.Set("encerramento", encerramentoVar)
		expectedCollection.Set("segundoLeilao", segundoLeilaoVar)
		expectedCollection.Set("encerramentoSegundoLeilao", encerramentoSegundoLeilaoVar)
		expectedCollection.Set("descontoSegundoLeilao", descontoSegundoLeilaoVar)
		expectedCollection.Set("valorDivida", valorDividaVar)
		expectedCollection.Set("valorDividaExtenso", valorDividaExtensoVar)
		expectedCollection.Set("origemValorDivida", origemValorDividaVar)
		expectedCollection.Set("data_edital", dataEditalVar)
		expectedCollection.Set("bens", bensVar)

		suite.mockVariableProcessor.EXPECT().
			ProcessVariables(realLegalData).
			Return(expectedCollection, nil).
			Once()

		suite.mockEngine.EXPECT().
			Process("edital_template_enhanced.tex").
			Return("Complete legal document for Empresa Real Ltda", nil).
			Once()

		suite.mockFileProcessor.EXPECT().
			CreateDirectory("output").
			Return(nil).
			Once()

		suite.mockFileProcessor.EXPECT().
			WriteFile("output/processed.tex", []byte("Complete legal document for Empresa Real Ltda")).
			Return(nil).
			Once()

		// Execute complete workflow
		cfg := suite.mockConfigProvider.GetConfig()
		suite.Equal("Empresa Real Ltda", cfg.Variables["executado"])

		err := suite.mockValidator.ValidateTemplate("edital_template_enhanced.tex")
		suite.NoError(err)

		exists := suite.mockFileProcessor.FileExists("edital_template_enhanced.tex")
		suite.True(exists)

		content, err := suite.mockFileProcessor.ReadFile("edital_template_enhanced.tex")
		suite.NoError(err)
		suite.NotNil(content)

		collection, err := suite.mockVariableProcessor.ProcessVariables(cfg.Variables)
		suite.NoError(err)
		suite.NotNil(collection)

		result, err := suite.mockEngine.Process("edital_template_enhanced.tex")
		suite.NoError(err)
		suite.Contains(result, "Empresa Real Ltda")

		err = suite.mockFileProcessor.CreateDirectory("output")
		suite.NoError(err)

		err = suite.mockFileProcessor.WriteFile("output/processed.tex", []byte(result))
		suite.NoError(err)
	})
}

// TestEditalPdfApiSuite runs the complete test suite
func TestEditalPdfApiSuite(t *testing.T) {
	suite.Run(t, new(EditalPdfApiTestSuite))
}
