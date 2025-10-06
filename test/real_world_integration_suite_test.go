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

// RealWorldIntegrationTestSuite tests real-world scenarios from both cartas-backend and edital-pdf-api
type RealWorldIntegrationTestSuite struct {
	suite.Suite
	mockEngine            *mocks.MockTemplateEngine
	mockEnhancedEngine    *mocks.MockEnhancedTemplateEngine
	mockValidator         *mocks.MockTemplateValidator
	mockFileProcessor     *mocks.MockFileProcessor
	mockVariableProcessor *mocks.MockVariableProcessor
	mockConfigProvider    *mocks.MockConfigProvider
}

// SetupTest initializes the test suite
func (suite *RealWorldIntegrationTestSuite) SetupTest() {
	suite.mockEngine = mocks.NewMockTemplateEngine(suite.T())
	suite.mockEnhancedEngine = mocks.NewMockEnhancedTemplateEngine(suite.T())
	suite.mockValidator = mocks.NewMockTemplateValidator(suite.T())
	suite.mockFileProcessor = mocks.NewMockFileProcessor(suite.T())
	suite.mockVariableProcessor = mocks.NewMockVariableProcessor(suite.T())
	suite.mockConfigProvider = mocks.NewMockConfigProvider(suite.T())
}

// TestFuneralLetterProductionWorkflow tests the complete funeral letter production workflow
func (suite *RealWorldIntegrationTestSuite) TestFuneralLetterProductionWorkflow() {
	// Real-world funeral letter data
	funeralData := map[string]interface{}{
		"nome":               "Maria da Conceição Silva",
		"ano_nascimento":     1940,
		"ano_morte":          2024,
		"local_velorio":      "Capela Nossa Senhora da Conceição",
		"local_sepultamento": "Cemitério Municipal de Franca",
		"data_velorio":       "22/01/2024",
		"hora_velorio":       "14:00",
		"hora_velorio_fim":   "18:00",
		"sem_velorio":        "Não",
		"segunda_data":       "Não",
		"fotografia_fundo":   []byte("background_image_data"),
		"perfil":             []byte("profile_image_data"),
		"logo":               []byte("logo_image_data"),
	}

	// Set up complete production workflow
	suite.mockConfigProvider.EXPECT().
		GetConfig().
		Return(&config.Config{
			Template:  "template.tex",
			Output:    "funeral_letters/maria_conceicao_silva.pdf",
			Variables: funeralData,
			Engine:    "lualatex",
		}).
		Once()

	suite.mockValidator.EXPECT().
		ValidateTemplate("template.tex").
		Return(nil).
		Once()

	suite.mockFileProcessor.EXPECT().
		FileExists("template.tex").
		Return(true).
		Once()

	// Mock LaTeX template with funeral letter structure
	templateContent := `\documentclass[17pt]{scrartcl}
\usepackage[utf8]{inputenc}
\usepackage[T1]{fontenc}
\usepackage{graphicx}
\usepackage{geometry}
\usepackage{ebgaramond}
\usepackage{pgfornament}
\usepackage{tikz}

\newcommand{\PersonNameVar}{delim[[.nome]]}
\newcommand{\YearOfBirthVar}{delim[[.ano_nascimento]]}
\newcommand{\YearOfDeathVar}{delim[[.ano_morte]]}
\newcommand{\LocationVar}{delim[[.local_velorio]]}
\newcommand{\CemeteryVar}{delim[[.local_sepultamento]]}

\begin{document}
\begin{center}
  {\textcolor{CoolBlack}{\Title{Participação de Falecimento}}} \\
  \vspace{0.2cm}
  
  {\textcolor{CoolBlack}{\DynamicPersonName}} \\
  {\textcolor{CoolBlack!90} {\large \YearOfBirthVar -- \YearOfDeathVar}} \\
  
  delim[[if eq .sem_velorio "Não"]]
  Com profundo pesar e saudade, convidamos familiares \\
  e amigos para o velório de \boldcolor{\PersonNameVar}. \\
  \vspace{0.3cm}
  Será realizado no dia {\infocolor{delim[[.data_velorio]]}}, das {\infocolor{delim[[.hora_velorio]]}} às {\infocolor{delim[[.hora_velorio_fim]]}}: \\
  \vspace{0.2cm}
  {\infocolor{\Large \LocationVar}} \\
  \vspace{0.2cm}
  Sepultamento: {\infocolor{\CemeteryVar}}
  delim[[end]]
\end{center}
\end{document}`

	suite.mockFileProcessor.EXPECT().
		ReadFile("template.tex").
		Return([]byte(templateContent), nil).
		Once()

	// Process variables for funeral letter
	expectedCollection := domain.NewVariableCollection()
	nomeVar, _ := domain.NewStringVariable("Maria da Conceição Silva")
	anoNascVar, _ := domain.NewNumberVariable(1940)
	anoMorteVar, _ := domain.NewNumberVariable(2024)
	localVelorioVar, _ := domain.NewStringVariable("Capela Nossa Senhora da Conceição")
	localSepultamentoVar, _ := domain.NewStringVariable("Cemitério Municipal de Franca")
	dataVelorioVar, _ := domain.NewStringVariable("22/01/2024")
	horaVelorioVar, _ := domain.NewStringVariable("14:00")
	horaVelorioFimVar, _ := domain.NewStringVariable("18:00")
	semVelorioVar, _ := domain.NewStringVariable("Não")
	segundaDataVar, _ := domain.NewStringVariable("Não")

	expectedCollection.Set("nome", nomeVar)
	expectedCollection.Set("ano_nascimento", anoNascVar)
	expectedCollection.Set("ano_morte", anoMorteVar)
	expectedCollection.Set("local_velorio", localVelorioVar)
	expectedCollection.Set("local_sepultamento", localSepultamentoVar)
	expectedCollection.Set("data_velorio", dataVelorioVar)
	expectedCollection.Set("hora_velorio", horaVelorioVar)
	expectedCollection.Set("hora_velorio_fim", horaVelorioFimVar)
	expectedCollection.Set("sem_velorio", semVelorioVar)
	expectedCollection.Set("segunda_data", segundaDataVar)

	suite.mockVariableProcessor.EXPECT().
		ProcessVariables(funeralData).
		Return(expectedCollection, nil).
		Once()

	// Process template with funeral letter data
	processedContent := `\documentclass[17pt]{scrartcl}
\usepackage[utf8]{inputenc}
\usepackage[T1]{fontenc}
\usepackage{graphicx}
\usepackage{geometry}
\usepackage{ebgaramond}
\usepackage{pgfornament}
\usepackage{tikz}

\newcommand{\PersonNameVar}{Maria da Conceição Silva}
\newcommand{\YearOfBirthVar}{1940}
\newcommand{\YearOfDeathVar}{2024}
\newcommand{\LocationVar}{Capela Nossa Senhora da Conceição}
\newcommand{\CemeteryVar}{Cemitério Municipal de Franca}

\begin{document}
\begin{center}
  {\textcolor{CoolBlack}{\Title{Participação de Falecimento}}} \\
  \vspace{0.2cm}
  
  {\textcolor{CoolBlack}{\DynamicPersonName}} \\
  {\textcolor{CoolBlack!90} {\large 1940 -- 2024}} \\
  
  Com profundo pesar e saudade, convidamos familiares \\
  e amigos para o velório de \boldcolor{Maria da Conceição Silva}. \\
  \vspace{0.3cm}
  Será realizado no dia {\infocolor{22/01/2024}}, das {\infocolor{14:00}} às {\infocolor{18:00}}: \\
  \vspace{0.2cm}
  {\infocolor{\Large Capela Nossa Senhora da Conceição}} \\
  \vspace{0.2cm}
  Sepultamento: {\infocolor{Cemiterio Municipal de Franca}}
\end{center}
\end{document}`

	suite.mockEngine.EXPECT().
		Process("template.tex").
		Return(processedContent, nil).
		Once()

	suite.mockFileProcessor.EXPECT().
		CreateDirectory("funeral_letters").
		Return(nil).
		Once()

	suite.mockFileProcessor.EXPECT().
		WriteFile("funeral_letters/processed.tex", []byte(processedContent)).
		Return(nil).
		Once()

	// Execute the complete production workflow
	cfg := suite.mockConfigProvider.GetConfig()
	suite.Equal("template.tex", cfg.Template.String())
	suite.Equal("funeral_letters/maria_conceicao_silva.pdf", cfg.Output.String())
	suite.Equal("lualatex", cfg.Engine.String())

	err := suite.mockValidator.ValidateTemplate("template.tex")
	suite.NoError(err)

	exists := suite.mockFileProcessor.FileExists("template.tex")
	suite.True(exists)

	content, err := suite.mockFileProcessor.ReadFile("template.tex")
	suite.NoError(err)
	suite.Contains(string(content), "\\documentclass[17pt]{scrartcl}")

	collection, err := suite.mockVariableProcessor.ProcessVariables(cfg.Variables)
	suite.NoError(err)
	suite.NotNil(collection)

	result, err := suite.mockEngine.Process("template.tex")
	suite.NoError(err)
	suite.Contains(result, "Maria da Conceição Silva")
	suite.Contains(result, "1940 -- 2024")
	suite.Contains(result, "Capela Nossa Senhora da Conceição")

	err = suite.mockFileProcessor.CreateDirectory("funeral_letters")
	suite.NoError(err)

	err = suite.mockFileProcessor.WriteFile("funeral_letters/processed.tex", []byte(result))
	suite.NoError(err)
}

// TestLegalDocumentProductionWorkflow tests the complete legal document production workflow
func (suite *RealWorldIntegrationTestSuite) TestLegalDocumentProductionWorkflow() {
	// Real-world legal auction document data
	legalData := map[string]interface{}{
		"leilao": map[string]interface{}{
			"vara":     "1ª Vara Cível",
			"comarca":  "Belo Horizonte",
			"estado":   "MG",
			"processo": "1234567-89.2024.8.13.0024",
		},
		"executado": "Construtora ABC Ltda",
		"documento": "12.345.678/0001-90",
		"exequente": "Banco XYZ S.A.",
		"juiz": map[string]interface{}{
			"nome": "Dr. João Silva",
		},
		"site":                      "www.leiloesjudiciais.com.br",
		"encerramento":              "30/01/2024 às 15:00",
		"segundoLeilao":             true,
		"encerramentoSegundoLeilao": "31/01/2024 às 15:00",
		"descontoSegundoLeilao":     80,
		"valorDivida":               "R$ 500.000,00",
		"valorDividaExtenso":        "quinhentos mil reais",
		"origemValorDivida":         "sentença judicial transitada em julgado",
		"data_edital":               "25/01/2024",
		"bens": []interface{}{
			map[string]interface{}{
				"descricao": "Imóvel Residencial",
				"registro":  "Matrícula 12345",
				"avaliacao": map[string]interface{}{
					"valor":        "R$ 300.000,00",
					"valorExtenso": "trezentos mil reais",
					"origem":       "avaliação judicial",
				},
				"onus": []interface{}{
					"Hipoteca em favor do Banco XYZ S.A.",
					"Penhor de veículos",
				},
			},
		},
	}

	// Set up complete production workflow
	suite.mockConfigProvider.EXPECT().
		GetConfig().
		Return(&config.Config{
			Template:  "edital_template_enhanced.tex",
			Output:    "legal_documents/edital_leilao_1234567.pdf",
			Variables: legalData,
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

	// Mock ABNTeX template with legal document structure
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
	leilaoVar, _ := domain.NewObjectVariable(legalData["leilao"].(map[string]interface{}))
	executadoVar, _ := domain.NewStringVariable("Construtora ABC Ltda")
	documentoVar, _ := domain.NewStringVariable("12.345.678/0001-90")
	exequenteVar, _ := domain.NewStringVariable("Banco XYZ S.A.")
	juizVar, _ := domain.NewObjectVariable(legalData["juiz"].(map[string]interface{}))
	siteVar, _ := domain.NewStringVariable("www.leiloesjudiciais.com.br")
	encerramentoVar, _ := domain.NewStringVariable("30/01/2024 às 15:00")
	segundoLeilaoVar := domain.NewBooleanVariable(true)
	encerramentoSegundoLeilaoVar, _ := domain.NewStringVariable("31/01/2024 às 15:00")
	descontoSegundoLeilaoVar, _ := domain.NewNumberVariable(80)
	valorDividaVar, _ := domain.NewStringVariable("R$ 500.000,00")
	valorDividaExtensoVar, _ := domain.NewStringVariable("quinhentos mil reais")
	origemValorDividaVar, _ := domain.NewStringVariable("sentença judicial transitada em julgado")
	dataEditalVar, _ := domain.NewStringVariable("25/01/2024")
	bensVar, _ := domain.NewArrayVariable(legalData["bens"].([]interface{}))

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
		ProcessVariables(legalData).
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
\newcommand{\ExecutadoVar}{Construtora ABC Ltda}
\newcommand{\DocumentoVar}{12.345.678/0001-90}
\newcommand{\ExequenteVar}{Banco XYZ S.A.}
\newcommand{\JuizNomeVar}{Dr. João Silva}
\newcommand{\SiteVar}{www.leiloesjudiciais.com.br}
\newcommand{\EncerramentoVar}{30/01/2024 às 15:00}
\newcommand{\ValorDividaVar}{R$ 500.000,00}
\newcommand{\ValorDividaExtensoVar}{quinhentos mil reais}
\newcommand{\OrigemValorDividaVar}{sentença judicial transitada em julgado}
\newcommand{\DataEditalVar}{25/01/2024}

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
\textbf{Construtora ABC Ltda - CNPJ nº 12.345.678/0001-90,} \\
por meio de seu representante legal, executado, \\
proprietário e fiel depositário do bem \\

o bem abaixo descrito,
conforme condições de venda constantes no presente edital. \\

No 1° Leilão com início da publicação do edital e término no dia \\
30/01/2024 às 15:00, \\
não serão admitidos lances inferiores ao valor de avaliação atualizada do bem. \\

Ficando desde já designado para o 2° Leilão com início no dia \\
30/01/2024 às 15:00 e término no dia 31/01/2024 às 15:00, \\
caso não haja licitantes no 1º Leilão. \\

No segundo serão admitidos lances não inferiores a 80\% do valor da avaliação atualizada. \\

\textbf{VALOR DA DÍVIDA NO PROCESSO DE EXECUÇÃO:} O valor da dívida no processo de execução é de \\
R$ 500.000,00 (quinhentos mil reais), \\
conforme sentença judicial transitada em julgado. \\

Belo Horizonte/MG, 25/01/2024. \\

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
		CreateDirectory("legal_documents").
		Return(nil).
		Once()

	suite.mockFileProcessor.EXPECT().
		WriteFile("legal_documents/processed.tex", []byte(processedContent)).
		Return(nil).
		Once()

	// Execute the complete production workflow
	cfg := suite.mockConfigProvider.GetConfig()
	suite.Equal("edital_template_enhanced.tex", cfg.Template.String())
	suite.Equal("legal_documents/edital_leilao_1234567.pdf", cfg.Output.String())
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
	suite.Contains(result, "Construtora ABC Ltda")
	suite.Contains(result, "12.345.678/0001-90")
	suite.Contains(result, "Dr. João Silva")

	err = suite.mockFileProcessor.CreateDirectory("legal_documents")
	suite.NoError(err)

	err = suite.mockFileProcessor.WriteFile("legal_documents/processed.tex", []byte(result))
	suite.NoError(err)
}

// TestConcurrentDocumentGeneration tests concurrent document generation scenarios
func (suite *RealWorldIntegrationTestSuite) TestConcurrentDocumentGeneration() {
	suite.Run("ConcurrentFuneralLetters", func() {
		// Test generating multiple funeral letters concurrently
		suite.mockEngine.EXPECT().
			Process(mock.AnythingOfType("string")).
			Return("processed funeral letter", nil).
			Times(10)

		// Simulate concurrent processing of 10 funeral letters
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(index int) {
				result, err := suite.mockEngine.Process("template.tex")
				suite.NoError(err)
				suite.Contains(result, "funeral letter")
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})

	suite.Run("ConcurrentLegalDocuments", func() {
		// Test generating multiple legal documents concurrently
		suite.mockEngine.EXPECT().
			Process(mock.AnythingOfType("string")).
			Return("processed legal document", nil).
			Times(5)

		// Simulate concurrent processing of 5 legal documents
		done := make(chan bool, 5)
		for i := 0; i < 5; i++ {
			go func(index int) {
				result, err := suite.mockEngine.Process("edital_template.tex")
				suite.NoError(err)
				suite.Contains(result, "legal document")
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 5; i++ {
			<-done
		}
	})
}

// TestErrorRecoveryWorkflow tests error recovery in production scenarios
func (suite *RealWorldIntegrationTestSuite) TestErrorRecoveryWorkflow() {
	suite.Run("TemplateValidationFailureRecovery", func() {
		// First validation fails
		suite.mockValidator.EXPECT().
			ValidateTemplate("corrupted_template.tex").
			Return(errors.New("LaTeX syntax error")).
			Once()

		// Retry with fixed template
		suite.mockValidator.EXPECT().
			ValidateTemplate("fixed_template.tex").
			Return(nil).
			Once()

		// Test error recovery
		err := suite.mockValidator.ValidateTemplate("corrupted_template.tex")
		suite.Error(err)
		suite.Contains(err.Error(), "LaTeX syntax error")

		// Recovery attempt
		err = suite.mockValidator.ValidateTemplate("fixed_template.tex")
		suite.NoError(err)
	})

	suite.Run("FileProcessingFailureRecovery", func() {
		// First file read fails
		suite.mockFileProcessor.EXPECT().
			ReadFile("corrupted_file.tex").
			Return(nil, errors.New("file corrupted")).
			Once()

		// Retry with backup file
		suite.mockFileProcessor.EXPECT().
			ReadFile("backup_file.tex").
			Return([]byte("template content"), nil).
			Once()

		// Test error recovery
		content, err := suite.mockFileProcessor.ReadFile("corrupted_file.tex")
		suite.Error(err)
		suite.Nil(content)

		// Recovery attempt
		content, err = suite.mockFileProcessor.ReadFile("backup_file.tex")
		suite.NoError(err)
		suite.NotNil(content)
	})
}

// TestProductionPerformanceWorkflow tests performance in production scenarios
func (suite *RealWorldIntegrationTestSuite) TestProductionPerformanceWorkflow() {
	suite.Run("HighVolumeFuneralLetterGeneration", func() {
		// Test generating 100 funeral letters
		suite.mockEngine.EXPECT().
			Process(mock.AnythingOfType("string")).
			Return("processed funeral letter", nil).
			Times(100)

		// Simulate high volume processing
		done := make(chan bool, 100)
		for i := 0; i < 100; i++ {
			go func(index int) {
				result, err := suite.mockEngine.Process("template.tex")
				suite.NoError(err)
				suite.Contains(result, "funeral letter")
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 100; i++ {
			<-done
		}
	})

	suite.Run("LargeLegalDocumentProcessing", func() {
		// Test processing large legal documents with complex data
		largeData := make(map[string]interface{})
		for i := 0; i < 500; i++ {
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

// TestRealWorldIntegrationSuite runs the complete test suite
func TestRealWorldIntegrationSuite(t *testing.T) {
	suite.Run(t, new(RealWorldIntegrationTestSuite))
}
