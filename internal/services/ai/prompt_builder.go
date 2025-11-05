package ai

import (
	"fmt"
	"strings"
	"time"
)

// PromptBuilder creates optimized prompts for Lithuanian grocery flyer analysis
type PromptBuilder struct {
	language     string
	storeContext map[string]string
	categories   []string
}

// NewPromptBuilder creates a new Lithuanian prompt builder
func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{
		language: "lithuanian",
		storeContext: map[string]string{
			"iki":    "IKI yra populiari Lietuvos prekybos tinklai",
			"maxima": "Maxima yra didžiausias maisto prekių tinklas Lietuvoje",
			"rimi":   "Rimi yra skandinavų prekybos tinklas veikiantis Lietuvoje",
		},
		categories: []string{
			"mėsa ir žuvis", "pieno produktai", "duona ir konditerija",
			"vaisiai ir daržovės", "gėrimai", "šaldyti produktai",
			"konservai", "kruopos ir makaronai", "saldumynai",
			"higienos prekės", "namų ūkio prekės", "alkoholiniai gėrimai",
		},
	}
}

// ProductExtractionPrompt creates a prompt for extracting products from flyer pages
func (pb *PromptBuilder) ProductExtractionPrompt(storeCode string, pageNumber int) string {
	storeContext := pb.getStoreContext(storeCode)

	prompt := fmt.Sprintf(`Analizuok šį %s prekybos tinklo leidinio %d puslapį ir ištrauk visų produktų informaciją.

KONTEKSTAS:
%s

UŽDUOTIS:
Ištrauk visus aiškiai matomus produktus su kainomis. Kiekvienam produktui nurodyk:

1. Produkto pavadinimas (lietuviškai, kaip parašyta leidinyje)
2. Kaina (su valiuta, pvz., "2,99 €")
3. Mato vienetas/kiekis (pvz., "1 kg", "500 g", "1 l", "vnt.")
4. Nuolaidos informacija (jei yra)
5. Prekės ženklas (jei matomas)
6. Kategorija (iš šių: %s)

FORMATAS:
Grąžink JSON masyvą su objektais:
{
  "products": [
    {
      "name": "produkto pavadinimas",
      "price": "kaina su valiuta",
      "unit": "mato vienetas",
      "original_price": "pradinė kaina jei yra nuolaida",
      "discount": "nuolaidos aprašymas",
      "brand": "prekės ženklas",
      "category": "kategorija"
    }
  ]
}

SVARBU:
- Ištrauk tik produktus su aiškiai matomais kainomis
- Išlaikyk originalų lietuvišką tekstą
- Jei kaina neaiški, neįtraukk produkto
- Už vienetus naudok standartines santrumpas (kg, g, l, ml, vnt.)
- Tiksliai nuraidyk kainų formatus (pvz., "1,99 €", "2.50 €")`,
		strings.ToUpper(storeCode), pageNumber, storeContext, strings.Join(pb.categories, ", "))

	return prompt
}

// TextExtractionPrompt creates a prompt for extracting all text from a flyer page
func (pb *PromptBuilder) TextExtractionPrompt(storeCode string) string {
	storeContext := pb.getStoreContext(storeCode)

	prompt := fmt.Sprintf(`Ištrauk visą tekstą iš šio %s prekybos tinklo leidinio puslapio.

KONTEKSTAS:
%s

UŽDUOTIS:
Ištrauk ir struktūrizuok visą matomą tekstą, įskaitant:

1. Produktų pavadinimus
2. Kainas ir nuolaidas
3. Kampanijų aprašymus
4. Datų informaciją
5. Kontaktinę informaciją
6. Kitus svarbius tekstus

FORMATAS:
Grąžink JSON objektą:
{
  "header_text": "antraštės ir svarbus tekstas",
  "products_text": "produktų tekstas",
  "prices_text": "kainų informacija",
  "dates_text": "datos ir galiojimo laikas",
  "promotional_text": "akcijų ir nuolaidų tekstas",
  "other_text": "kitas tekstas"
}

SVARBU:
- Išlaikyk originalų lietuvišką rašybą ir skyrybą
- Struktūrizuok panašų turinį į atitinkamus laukus
- Nekoreguok kainų formatų
- Ištrauk visą matomą tekstą, net jei jis atrodo nesvarus`,
		strings.ToUpper(storeCode), storeContext)

	return prompt
}

// ValidationPrompt creates a prompt for validating extracted product data
func (pb *PromptBuilder) ValidationPrompt(extractedData string) string {
	prompt := `Patikrink ir pakoreguok ištrauktų produktų duomenis.

GAUTAS DUOMENYS:
%s

TIKRINIMO KRITERIJAI:
1. Kainų formatai turi būti teisingi (pvz., "1,99 €", "2.50 €")
2. Produktų pavadinimai turi būti lietuviškai
3. Mato vienetai turi būti standartiniai (kg, g, l, ml, vnt.)
4. Kategorijos turi atitikti šiuos: %s
5. Nuolaidos informacija turi būti aiški

UŽDUOTIS:
Pataisyk klaidingas kainas, pavadinimus ir kategorijas. Pašalink produktus be kainų.

FORMATAS:
Grąžink pataisytą JSON su papildomu lauku "validation_notes" kiekvienam produktui:
{
  "products": [
    {
      "name": "pataisytas pavadinimas",
      "price": "pataisyta kaina",
      "unit": "pataisytas vienetas",
      "original_price": "pradinė kaina",
      "discount": "nuolaidos aprašymas",
      "brand": "prekės ženklas",
      "category": "pataisyta kategorija",
      "validation_notes": "koregavimo pastabos jei būtina"
    }
  ],
  "removed_products": [
    {
      "reason": "pašalinimo priežastis",
      "original_data": "originalūs duomenys"
    }
  ]
}`

	return fmt.Sprintf(prompt, extractedData, strings.Join(pb.categories, ", "))
}

// CategoryClassificationPrompt creates a prompt for classifying product categories
func (pb *PromptBuilder) CategoryClassificationPrompt(productName string) string {
	prompt := fmt.Sprintf(`Klasifikuok šį produktą pagal kategoriją.

PRODUKTO PAVADINIMAS: "%s"

GALIMOS KATEGORIJOS:
%s

UŽDUOTIS:
Nustatyk tinkamiausią kategoriją šiam produktui. Atsižvelgk į:
- Produkto pobūdį ir paskirtį
- Įprastą kategorizaciją prekybos centruose
- Lietuvių kalbos specifiką

FORMATAS:
Grąžink JSON objektą:
{
  "category": "pasirinkta kategorija",
  "confidence": 0.95,
  "reasoning": "argumentacija lietuviškai"
}

SVARBU:
- Pasirink tik iš pateiktų kategorijų
- Nurodyk pasitikėjimo lygį (0.0-1.0)
- Paaiškink pasirinkimą lietuviškai`,
		productName, strings.Join(pb.categories, "\n- "))

	return prompt
}

// PriceAnalysisPrompt creates a prompt for analyzing pricing information
func (pb *PromptBuilder) PriceAnalysisPrompt(products []string) string {
	productsText := strings.Join(products, "\n")

	prompt := fmt.Sprintf(`Analizuok kainų informaciją šiuose produktuose.

PRODUKTAI:
%s

ANALIZĖS KRITERIJAI:
1. Kainų formatų nuoseklumas
2. Nuolaidų apskaičiavimas
3. Kainų palyginimas pagal vienetą
4. Akcijų galiojimo datos

UŽDUOTIS:
Patikrink kainų teisingumą ir apskaičiuok papildomą informaciją.

FORMATAS:
Grąžink JSON objektą:
{
  "pricing_analysis": {
    "total_products": skaičius,
    "products_on_sale": skaičius,
    "average_discount_percentage": procentai,
    "price_range": {
      "min": "minimali kaina",
      "max": "maksimali kaina"
    }
  },
  "price_corrections": [
    {
      "product": "produkto pavadinimas",
      "original_price": "originali kaina",
      "corrected_price": "pataisyta kaina",
      "reason": "pataisymo priežastis"
    }
  ],
  "unit_price_calculations": [
    {
      "product": "produkto pavadinimas",
      "price_per_unit": "kaina už vienetą",
      "comparison_note": "palyginimo pastaba"
    }
  ]
}`,
		productsText)

	return prompt
}

// getStoreContext returns context information for a specific store
func (pb *PromptBuilder) getStoreContext(storeCode string) string {
	if context, exists := pb.storeContext[strings.ToLower(storeCode)]; exists {
		return context
	}
	return "Lietuvos prekybos tinklas"
}

// LayoutAnalysisPrompt creates a prompt for analyzing page layout
func (pb *PromptBuilder) LayoutAnalysisPrompt() string {
	return `Analizuok šio leidinio puslapio išdėstymą ir struktūrą.

UŽDUOTIS:
Identifikuok ir apibūdink:

1. Puslapio struktūrą (antraštės, skyriai, kolonos)
2. Produktų išdėstymą (eilės, grupės, kategorijos)
3. Kainų pozicionavimą
4. Akcijų ir nuolaidų išryškinimą
5. Prekių ženklų ir kategorijų žymėjimą

FORMATAS:
Grąžink JSON objektą:
{
  "layout_analysis": {
    "page_structure": "puslapio struktūros aprašymas",
    "product_layout": "produktų išdėstymo aprašymas",
    "sections": [
      {
        "type": "skyriaus tipas",
        "content": "turinio aprašymas",
        "position": "pozicija puslapyje"
      }
    ],
    "special_offers": [
      {
        "type": "akcijos tipas",
        "description": "akcijos aprašymas",
        "visual_emphasis": "vizualinio išryškinimo aprašymas"
      }
    ]
  }
}`
}

// QualityCheckPrompt creates a prompt for quality checking extracted data
func (pb *PromptBuilder) QualityCheckPrompt(extractedData string, originalImage string) string {
	prompt := fmt.Sprintf(`Patikrink ištrauktų duomenų kokybę palyginti su originaliu vaizdu.

IŠTRAUKTI DUOMENYS:
%s

KOKYBĖS KRITERIJAI:
1. Duomenų išsamumas (ar visi produktai ištraukti?)
2. Tikslumas (ar kainos ir pavadinimai teisingi?)
3. Nuoseklumas (ar formatai vienodi?)
4. Lietuvių kalbos teisingumas

UŽDUOTIS:
Įvertink duomenų kokybę ir pateik rekomendacijas.

FORMATAS:
Grąžink JSON objektą:
{
  "quality_score": 0.85,
  "completeness": 0.90,
  "accuracy": 0.80,
  "consistency": 0.85,
  "issues_found": [
    {
      "type": "problemos tipas",
      "description": "problemos aprašymas",
      "severity": "high/medium/low",
      "suggestion": "siūlymas sprendimui"
    }
  ],
  "missing_products": skaičius,
  "recommendations": [
    "rekomendacija 1",
    "rekomendacija 2"
  ]
}`,
		extractedData)

	return prompt
}

// BuildCustomPrompt creates a custom prompt with Lithuanian context
func (pb *PromptBuilder) BuildCustomPrompt(task, context, requirements string) string {
	timestamp := time.Now().Format("2006-01-02 15:04")

	prompt := fmt.Sprintf(`UŽDUOTIS: %s

KONTEKSTAS:
%s

REIKALAVIMAI:
%s

KALBOS SPECIFIKA:
- Visas tekstas turi būti lietuvių kalba
- Išlaikyk originalų rašybą ir skyrybą
- Naudok standartines lietuvių kalbos formas
- Prekių pavadinimus palik kaip originaliai parašyta

LAIKO ŽYMA: %s

Atlikk užduotį tiksliai pagal pateiktus reikalavimus.`,
		task, context, requirements, timestamp)

	return prompt
}

// GetAvailableCategories returns the list of available product categories
func (pb *PromptBuilder) GetAvailableCategories() []string {
	return pb.categories
}

// AddStoreContext adds context information for a new store
func (pb *PromptBuilder) AddStoreContext(storeCode, context string) {
	pb.storeContext[strings.ToLower(storeCode)] = context
}

// GetSupportedStores returns the list of stores with context
func (pb *PromptBuilder) GetSupportedStores() []string {
	stores := make([]string, 0, len(pb.storeContext))
	for store := range pb.storeContext {
		stores = append(stores, store)
	}
	return stores
}
