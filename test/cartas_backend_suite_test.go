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

// CartasBackendTestSuite tests AutoPDF functionality for funeral letter generation
type CartasBackendTestSuite struct {
	suite.Suite
	mockEngine            *mocks.MockTemplateEngine
	mockEnhancedEngine    *mocks.MockEnhancedTemplateEngine
	mockValidator         *mocks.MockTemplateValidator
	mockFileProcessor     *mocks.MockFileProcessor
	mockVariableProcessor *mocks.MockVariableProcessor
	mockConfigProvider    *mocks.MockConfigProvider
}

// SetupTest initializes the test suite
func (suite *CartasBackendTestSuite) SetupTest() {
	suite.mockEngine = mocks.NewMockTemplateEngine(suite.T())
	suite.mockEnhancedEngine = mocks.NewMockEnhancedTemplateEngine(suite.T())
	suite.mockValidator = mocks.NewMockTemplateValidator(suite.T())
	suite.mockFileProcessor = mocks.NewMockFileProcessor(suite.T())
	suite.mockVariableProcessor = mocks.NewMockVariableProcessor(suite.T())
	suite.mockConfigProvider = mocks.NewMockConfigProvider(suite.T())
}

// TestFuneralLetterGeneration tests the complete funeral letter generation workflow
func (suite *CartasBackendTestSuite) TestFuneralLetterGeneration() {
	// Test data representing a funeral letter entity
	letterData := map[string]interface{}{
		"nome":               "João da Silva",
		"ano_nascimento":     1950,
		"ano_morte":          2024,
		"local_velorio":      "Capela São José",
		"local_sepultamento": "Cemitério Municipal",
		"data_velorio":       "15/01/2024",
		"hora_velorio":       "14:00",
		"hora_velorio_fim":   "18:00",
		"sem_velorio":        "Não",
		"segunda_data":       "Não",
		"fotografia_fundo":   "background.jpg",
		"perfil":             "profile.jpg",
		"logo":               "logo.png",
	}

	// Set up expectations for the complete workflow
	suite.mockConfigProvider.EXPECT().
		GetConfig().
		Return(&config.Config{
			Template:  "template.tex",
			Output:    "funeral_letter.pdf",
			Variables: letterData,
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

	// Mock template content with funeral letter specific variables
	templateContent := `\documentclass[17pt]{scrartcl}
\usepackage[utf8]{inputenc}
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
	nomeVar, _ := domain.NewStringVariable("João da Silva")
	anoNascVar, _ := domain.NewNumberVariable(1950)
	anoMorteVar, _ := domain.NewNumberVariable(2024)
	localVelorioVar, _ := domain.NewStringVariable("Capela São José")
	localSepultamentoVar, _ := domain.NewStringVariable("Cemitério Municipal")
	dataVelorioVar, _ := domain.NewStringVariable("15/01/2024")
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
		ProcessVariables(letterData).
		Return(expectedCollection, nil).
		Once()

	// Process template with funeral letter data
	processedContent := `\documentclass[17pt]{scrartcl}
\usepackage[utf8]{inputenc}
\usepackage{graphicx}
\usepackage{geometry}
\usepackage{ebgaramond}
\usepackage{pgfornament}
\usepackage{tikz}

\newcommand{\PersonNameVar}{João da Silva}
\newcommand{\YearOfBirthVar}{1950}
\newcommand{\YearOfDeathVar}{2024}
\newcommand{\LocationVar}{Capela São José}
\newcommand{\CemeteryVar}{Cemitério Municipal}

\begin{document}
\begin{center}
  {\textcolor{CoolBlack}{\Title{Participação de Falecimento}}} \\
  \vspace{0.2cm}
  
  {\textcolor{CoolBlack}{\DynamicPersonName}} \\
  {\textcolor{CoolBlack!90} {\large 1950 -- 2024}} \\
  
  Com profundo pesar e saudade, convidamos familiares \\
  e amigos para o velório de \boldcolor{João da Silva}. \\
  \vspace{0.3cm}
  Será realizado no dia {\infocolor{15/01/2024}}, das {\infocolor{14:00}} às {\infocolor{18:00}}: \\
  \vspace{0.2cm}
  {\infocolor{\Large Capela São José}} \\
  \vspace{0.2cm}
  Sepultamento: {\infocolor{Cemiterio Municipal}}
\end{center}
\end{document}`

	suite.mockEngine.EXPECT().
		Process("template.tex").
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
	suite.Equal("template.tex", cfg.Template.String())
	suite.Equal("funeral_letter.pdf", cfg.Output.String())
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
	suite.Contains(result, "João da Silva")
	suite.Contains(result, "1950 -- 2024")
	suite.Contains(result, "Capela São José")

	err = suite.mockFileProcessor.CreateDirectory("output")
	suite.NoError(err)

	err = suite.mockFileProcessor.WriteFile("output/processed.tex", []byte(result))
	suite.NoError(err)
}

// TestFuneralLetterWithNoWake tests funeral letter generation for direct burial
func (suite *CartasBackendTestSuite) TestFuneralLetterWithNoWake() {
	letterData := map[string]interface{}{
		"nome":                     "Maria Santos",
		"ano_nascimento":           1945,
		"ano_morte":                2024,
		"local_sepultamento":       "Cemitério da Consolação",
		"sem_velorio":              "Sim",
		"data_sepultamento_direto": "16/01/2024",
		"hora_sepultamento_direto": "10:00",
		"fotografia_fundo":         "background.jpg",
		"perfil":                   "profile.jpg",
		"logo":                     "logo.png",
	}

	suite.mockConfigProvider.EXPECT().
		GetConfig().
		Return(&config.Config{
			Template:  "template.tex",
			Output:    "funeral_letter_no_wake.pdf",
			Variables: letterData,
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

	suite.mockFileProcessor.EXPECT().
		ReadFile("template.tex").
		Return([]byte("template content"), nil).
		Once()

	expectedCollection := domain.NewVariableCollection()
	nomeVar, _ := domain.NewStringVariable("Maria Santos")
	anoNascVar, _ := domain.NewNumberVariable(1945)
	anoMorteVar, _ := domain.NewNumberVariable(2024)
	localSepultamentoVar, _ := domain.NewStringVariable("Cemitério da Consolação")
	semVelorioVar, _ := domain.NewStringVariable("Sim")
	dataSepultamentoVar, _ := domain.NewStringVariable("16/01/2024")
	horaSepultamentoVar, _ := domain.NewStringVariable("10:00")

	expectedCollection.Set("nome", nomeVar)
	expectedCollection.Set("ano_nascimento", anoNascVar)
	expectedCollection.Set("ano_morte", anoMorteVar)
	expectedCollection.Set("local_sepultamento", localSepultamentoVar)
	expectedCollection.Set("sem_velorio", semVelorioVar)
	expectedCollection.Set("data_sepultamento_direto", dataSepultamentoVar)
	expectedCollection.Set("hora_sepultamento_direto", horaSepultamentoVar)

	suite.mockVariableProcessor.EXPECT().
		ProcessVariables(letterData).
		Return(expectedCollection, nil).
		Once()

	suite.mockEngine.EXPECT().
		Process("template.tex").
		Return("Processed content for direct burial", nil).
		Once()

	suite.mockFileProcessor.EXPECT().
		CreateDirectory("output").
		Return(nil).
		Once()

	suite.mockFileProcessor.EXPECT().
		WriteFile("output/processed.tex", []byte("Processed content for direct burial")).
		Return(nil).
		Once()

	// Execute workflow
	cfg := suite.mockConfigProvider.GetConfig()
	suite.Equal("Sim", cfg.Variables["sem_velorio"])

	err := suite.mockValidator.ValidateTemplate("template.tex")
	suite.NoError(err)

	exists := suite.mockFileProcessor.FileExists("template.tex")
	suite.True(exists)

	content, err := suite.mockFileProcessor.ReadFile("template.tex")
	suite.NoError(err)
	suite.NotNil(content)

	collection, err := suite.mockVariableProcessor.ProcessVariables(cfg.Variables)
	suite.NoError(err)
	suite.NotNil(collection)

	result, err := suite.mockEngine.Process("template.tex")
	suite.NoError(err)
	suite.Contains(result, "direct burial")

	err = suite.mockFileProcessor.CreateDirectory("output")
	suite.NoError(err)

	err = suite.mockFileProcessor.WriteFile("output/processed.tex", []byte(result))
	suite.NoError(err)
}

// TestFuneralLetterWithTwoDays tests funeral letter with two-day wake
func (suite *CartasBackendTestSuite) TestFuneralLetterWithTwoDays() {
	letterData := map[string]interface{}{
		"nome":                   "Carlos Oliveira",
		"ano_nascimento":         1960,
		"ano_morte":              2024,
		"local_velorio":          "Capela Nossa Senhora",
		"local_sepultamento":     "Cemitério São João",
		"data_velorio":           "17/01/2024",
		"hora_velorio":           "19:00",
		"hora_velorio_fim":       "22:00",
		"segunda_data":           "Sim",
		"data_velorio_final":     "18/01/2024",
		"hora_velorio_final":     "08:00",
		"hora_velorio_final_fim": "12:00",
		"sem_velorio":            "Não",
		"fotografia_fundo":       "background.jpg",
		"perfil":                 "profile.jpg",
		"logo":                   "logo.png",
	}

	suite.mockConfigProvider.EXPECT().
		GetConfig().
		Return(&config.Config{
			Template:  "template.tex",
			Output:    "funeral_letter_two_days.pdf",
			Variables: letterData,
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

	suite.mockFileProcessor.EXPECT().
		ReadFile("template.tex").
		Return([]byte("template content"), nil).
		Once()

	expectedCollection := domain.NewVariableCollection()
	nomeVar, _ := domain.NewStringVariable("Carlos Oliveira")
	anoNascVar, _ := domain.NewNumberVariable(1960)
	anoMorteVar, _ := domain.NewNumberVariable(2024)
	localVelorioVar, _ := domain.NewStringVariable("Capela Nossa Senhora")
	localSepultamentoVar, _ := domain.NewStringVariable("Cemitério São João")

	expectedCollection.Set("nome", nomeVar)
	expectedCollection.Set("ano_nascimento", anoNascVar)
	expectedCollection.Set("ano_morte", anoMorteVar)
	expectedCollection.Set("local_velorio", localVelorioVar)
	expectedCollection.Set("local_sepultamento", localSepultamentoVar)
	dataVelorioVar, _ := domain.NewStringVariable("17/01/2024")
	horaVelorioVar, _ := domain.NewStringVariable("19:00")
	horaVelorioFimVar, _ := domain.NewStringVariable("22:00")
	segundaDataVar, _ := domain.NewStringVariable("Sim")
	dataVelorioFinalVar, _ := domain.NewStringVariable("18/01/2024")

	expectedCollection.Set("data_velorio", dataVelorioVar)
	expectedCollection.Set("hora_velorio", horaVelorioVar)
	expectedCollection.Set("hora_velorio_fim", horaVelorioFimVar)
	expectedCollection.Set("segunda_data", segundaDataVar)
	expectedCollection.Set("data_velorio_final", dataVelorioFinalVar)
	horaVelorioFinalVar, _ := domain.NewStringVariable("08:00")
	horaVelorioFinalFimVar, _ := domain.NewStringVariable("12:00")
	semVelorioVar, _ := domain.NewStringVariable("Não")

	expectedCollection.Set("hora_velorio_final", horaVelorioFinalVar)
	expectedCollection.Set("hora_velorio_final_fim", horaVelorioFinalFimVar)
	expectedCollection.Set("sem_velorio", semVelorioVar)

	suite.mockVariableProcessor.EXPECT().
		ProcessVariables(letterData).
		Return(expectedCollection, nil).
		Once()

	suite.mockEngine.EXPECT().
		Process("template.tex").
		Return("Processed content for two-day wake", nil).
		Once()

	suite.mockFileProcessor.EXPECT().
		CreateDirectory("output").
		Return(nil).
		Once()

	suite.mockFileProcessor.EXPECT().
		WriteFile("output/processed.tex", []byte("Processed content for two-day wake")).
		Return(nil).
		Once()

	// Execute workflow
	cfg := suite.mockConfigProvider.GetConfig()
	suite.Equal("Sim", cfg.Variables["segunda_data"])

	err := suite.mockValidator.ValidateTemplate("template.tex")
	suite.NoError(err)

	exists := suite.mockFileProcessor.FileExists("template.tex")
	suite.True(exists)

	content, err := suite.mockFileProcessor.ReadFile("template.tex")
	suite.NoError(err)
	suite.NotNil(content)

	collection, err := suite.mockVariableProcessor.ProcessVariables(cfg.Variables)
	suite.NoError(err)
	suite.NotNil(collection)

	result, err := suite.mockEngine.Process("template.tex")
	suite.NoError(err)
	suite.Contains(result, "two-day wake")

	err = suite.mockFileProcessor.CreateDirectory("output")
	suite.NoError(err)

	err = suite.mockFileProcessor.WriteFile("output/processed.tex", []byte(result))
	suite.NoError(err)
}

// TestFuneralLetterErrorHandling tests error scenarios in funeral letter generation
func (suite *CartasBackendTestSuite) TestFuneralLetterErrorHandling() {
	suite.Run("TemplateValidationFailure", func() {
		suite.mockValidator.EXPECT().
			ValidateTemplate("invalid_template.tex").
			Return(errors.New("LaTeX syntax error: missing \\begin{document}")).
			Once()

		err := suite.mockValidator.ValidateTemplate("invalid_template.tex")
		suite.Error(err)
		suite.Contains(err.Error(), "LaTeX syntax error")
	})

	suite.Run("MissingTemplateFile", func() {
		suite.mockFileProcessor.EXPECT().
			FileExists("missing_template.tex").
			Return(false).
			Once()

		exists := suite.mockFileProcessor.FileExists("missing_template.tex")
		suite.False(exists)
	})

	suite.Run("FileReadError", func() {
		suite.mockFileProcessor.EXPECT().
			ReadFile("corrupted_template.tex").
			Return(nil, errors.New("permission denied")).
			Once()

		content, err := suite.mockFileProcessor.ReadFile("corrupted_template.tex")
		suite.Error(err)
		suite.Nil(content)
	})

	suite.Run("TemplateProcessingError", func() {
		suite.mockEngine.EXPECT().
			Process("error_template.tex").
			Return("", errors.New("template processing failed: undefined variable 'missing_var'")).
			Once()

		result, err := suite.mockEngine.Process("error_template.tex")
		suite.Error(err)
		suite.Empty(result)
		suite.Contains(err.Error(), "undefined variable")
	})
}

// TestFuneralLetterPerformance tests performance scenarios for funeral letter generation
func (suite *CartasBackendTestSuite) TestFuneralLetterPerformance() {
	suite.Run("HighVolumeLetterGeneration", func() {
		// Test generating multiple funeral letters concurrently
		suite.mockEngine.EXPECT().
			Process(mock.AnythingOfType("string")).
			Return("processed funeral letter", nil).
			Times(50)

		// Simulate concurrent processing of 50 funeral letters
		done := make(chan bool, 50)
		for i := 0; i < 50; i++ {
			go func(index int) {
				result, err := suite.mockEngine.Process("template.tex")
				suite.NoError(err)
				suite.Contains(result, "funeral letter")
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 50; i++ {
			<-done
		}
	})

	suite.Run("LargeImageProcessing", func() {
		// Test processing funeral letters with large images
		largeImageData := make([]byte, 10*1024*1024) // 10MB image
		letterData := map[string]interface{}{
			"nome":             "João da Silva",
			"fotografia_fundo": largeImageData,
			"perfil":           largeImageData,
		}

		suite.mockVariableProcessor.EXPECT().
			ProcessVariables(letterData).
			Return(domain.NewVariableCollection(), nil).
			Once()

		suite.mockEngine.EXPECT().
			Process("template.tex").
			Return("processed with large images", nil).
			Once()

		collection, err := suite.mockVariableProcessor.ProcessVariables(letterData)
		suite.NoError(err)
		suite.NotNil(collection)

		result, err := suite.mockEngine.Process("template.tex")
		suite.NoError(err)
		suite.Contains(result, "large images")
	})
}

// TestFuneralLetterIntegration tests integration scenarios
func (suite *CartasBackendTestSuite) TestFuneralLetterIntegration() {
	suite.Run("CompleteWorkflowWithRealData", func() {
		// Simulate real funeral letter data structure
		realLetterData := map[string]interface{}{
			"nome":               "Ana Maria da Silva",
			"ano_nascimento":     1935,
			"ano_morte":          2024,
			"local_velorio":      "Capela São Francisco",
			"local_sepultamento": "Cemitério Municipal de Franca",
			"data_velorio":       "20/01/2024",
			"hora_velorio":       "14:00",
			"hora_velorio_fim":   "18:00",
			"sem_velorio":        "Não",
			"segunda_data":       "Não",
			"fotografia_fundo":   []byte("background_image_data"),
			"perfil":             []byte("profile_image_data"),
			"logo":               []byte("logo_image_data"),
		}

		suite.mockConfigProvider.EXPECT().
			GetConfig().
			Return(&config.Config{
				Template:  "template.tex",
				Output:    "ana_maria_funeral_letter.pdf",
				Variables: realLetterData,
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

		suite.mockFileProcessor.EXPECT().
			ReadFile("template.tex").
			Return([]byte("template content"), nil).
			Once()

		expectedCollection := domain.NewVariableCollection()
		nomeVar, _ := domain.NewStringVariable("Ana Maria da Silva")
		anoNascVar, _ := domain.NewNumberVariable(1935)
		anoMorteVar, _ := domain.NewNumberVariable(2024)
		localVelorioVar, _ := domain.NewStringVariable("Capela São Francisco")
		localSepultamentoVar, _ := domain.NewStringVariable("Cemitério Municipal de Franca")
		dataVelorioVar, _ := domain.NewStringVariable("20/01/2024")
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
			ProcessVariables(realLetterData).
			Return(expectedCollection, nil).
			Once()

		suite.mockEngine.EXPECT().
			Process("template.tex").
			Return("Complete funeral letter for Ana Maria da Silva", nil).
			Once()

		suite.mockFileProcessor.EXPECT().
			CreateDirectory("output").
			Return(nil).
			Once()

		suite.mockFileProcessor.EXPECT().
			WriteFile("output/processed.tex", []byte("Complete funeral letter for Ana Maria da Silva")).
			Return(nil).
			Once()

		// Execute complete workflow
		cfg := suite.mockConfigProvider.GetConfig()
		suite.Equal("Ana Maria da Silva", cfg.Variables["nome"])

		err := suite.mockValidator.ValidateTemplate("template.tex")
		suite.NoError(err)

		exists := suite.mockFileProcessor.FileExists("template.tex")
		suite.True(exists)

		content, err := suite.mockFileProcessor.ReadFile("template.tex")
		suite.NoError(err)
		suite.NotNil(content)

		collection, err := suite.mockVariableProcessor.ProcessVariables(cfg.Variables)
		suite.NoError(err)
		suite.NotNil(collection)

		result, err := suite.mockEngine.Process("template.tex")
		suite.NoError(err)
		suite.Contains(result, "Ana Maria da Silva")

		err = suite.mockFileProcessor.CreateDirectory("output")
		suite.NoError(err)

		err = suite.mockFileProcessor.WriteFile("output/processed.tex", []byte(result))
		suite.NoError(err)
	})
}

// TestFuneralLetterEdgeCases tests edge cases in funeral letter generation
func (suite *CartasBackendTestSuite) TestFuneralLetterEdgeCases() {
	suite.Run("EmptyName", func() {
		letterData := map[string]interface{}{
			"nome": "",
		}

		suite.mockVariableProcessor.EXPECT().
			ProcessVariables(letterData).
			Return(domain.NewVariableCollection(), nil).
			Once()

		collection, err := suite.mockVariableProcessor.ProcessVariables(letterData)
		suite.NoError(err)
		suite.NotNil(collection)
	})

	suite.Run("VeryLongName", func() {
		longName := "João da Silva Santos Oliveira Pereira Rodrigues Costa Ferreira Almeida"
		letterData := map[string]interface{}{
			"nome": longName,
		}

		suite.mockVariableProcessor.EXPECT().
			ProcessVariables(letterData).
			Return(domain.NewVariableCollection(), nil).
			Once()

		collection, err := suite.mockVariableProcessor.ProcessVariables(letterData)
		suite.NoError(err)
		suite.NotNil(collection)
	})

	suite.Run("SpecialCharactersInName", func() {
		letterData := map[string]interface{}{
			"nome": "José da Silva & Filhos Ltda.",
		}

		suite.mockVariableProcessor.EXPECT().
			ProcessVariables(letterData).
			Return(domain.NewVariableCollection(), nil).
			Once()

		collection, err := suite.mockVariableProcessor.ProcessVariables(letterData)
		suite.NoError(err)
		suite.NotNil(collection)
	})
}

// TestCartasBackendSuite runs the complete test suite
func TestCartasBackendSuite(t *testing.T) {
	suite.Run(t, new(CartasBackendTestSuite))
}
