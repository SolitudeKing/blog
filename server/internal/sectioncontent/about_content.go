package sectioncontent

const (
	// AboutKickerMaxRunes caps the small label rendered above the about hero
	// heading (e.g. "About the keeper").
	AboutKickerMaxRunes = 32
	// AboutHeadingMaxRunes caps the heading rendered inside the about hero.
	AboutHeadingMaxRunes = 80
	// AboutLeadMaxRunes caps the multi-line lead paragraph rendered below
	// the about heading.
	AboutLeadMaxRunes = 240
	// AboutSignatureMaxRunes caps the short signature line rendered inside
	// the about hero.
	AboutSignatureMaxRunes = 80
	// AboutActionLabelMaxRunes caps the two short action button labels on
	// the about hero.
	AboutActionLabelMaxRunes = 16
	// AboutPrinciplesKickerMaxRunes caps the small label rendered above the
	// publishing principles heading.
	AboutPrinciplesKickerMaxRunes = 32
	// AboutPrinciplesHeadingMaxRunes caps the heading rendered above the
	// publishing principles list.
	AboutPrinciplesHeadingMaxRunes = 80
	// AboutPrinciplesIntroMaxRunes caps the multi-line description rendered
	// below the publishing principles heading.
	AboutPrinciplesIntroMaxRunes = 160
	// AboutPrincipleTitleMaxRunes caps the title of a single publishing
	// principle.
	AboutPrincipleTitleMaxRunes = 40
	// AboutPrincipleDescriptionMaxRunes caps the description of a single
	// publishing principle.
	AboutPrincipleDescriptionMaxRunes = 160
	// AboutContactKickerMaxRunes caps the small label rendered above the
	// contact section heading.
	AboutContactKickerMaxRunes = 32
	// AboutContactHeadingMaxRunes caps the contact section heading, both the
	// with-links and the empty variant.
	AboutContactHeadingMaxRunes = 80
	// AboutContactIntroMaxRunes caps the contact section description, both
	// the with-links and the empty variant.
	AboutContactIntroMaxRunes = 160
	// AboutContactEmptyCtaMaxRunes caps the call-to-action label shown when
	// no social links are configured.
	AboutContactEmptyCtaMaxRunes = 16
	// AboutPortraitLine1MaxRunes caps the first caption line rendered below
	// the about portrait.
	AboutPortraitLine1MaxRunes = 40
	// AboutPortraitLine2MaxRunes caps the second caption line rendered below
	// the about portrait.
	AboutPortraitLine2MaxRunes = 80
)

// AboutContent is the full set of theme-independent about page copy fields.
// Every field is required so the admin form and the public renderer can
// rely on a complete object.
type AboutContent struct {
	AboutKicker              string `json:"about_kicker"`
	AboutHeading             string `json:"about_heading"`
	AboutLead                string `json:"about_lead"`
	AboutSignature           string `json:"about_signature"`
	AboutContactLabel        string `json:"about_contact_label"`
	AboutReadingLabel        string `json:"about_reading_label"`
	AboutPrinciplesKicker    string `json:"about_principles_kicker"`
	AboutPrinciplesHeading   string `json:"about_principles_heading"`
	AboutPrinciplesIntro     string `json:"about_principles_intro"`
	Principle1Title          string `json:"principle_1_title"`
	Principle1Description    string `json:"principle_1_description"`
	Principle2Title          string `json:"principle_2_title"`
	Principle2Description    string `json:"principle_2_description"`
	Principle3Title          string `json:"principle_3_title"`
	Principle3Description    string `json:"principle_3_description"`
	AboutContactKicker       string `json:"about_contact_kicker"`
	AboutContactHeadingWith  string `json:"about_contact_heading_with"`
	AboutContactHeadingEmpty string `json:"about_contact_heading_empty"`
	AboutContactIntroWith    string `json:"about_contact_intro_with"`
	AboutContactIntroEmpty   string `json:"about_contact_intro_empty"`
	AboutContactEmptyCta     string `json:"about_contact_empty_cta"`
	AboutPortraitLine1       string `json:"about_portrait_line1"`
	AboutPortraitLine2       string `json:"about_portrait_line2"`
}

// DefaultAboutContent returns a complete, independently allocated copy of
// the shipped about page copy. The values match the strings that used to be
// hard-coded in AboutPage.vue so the migration is byte-for-byte identical
// when no admin override is set, except for the hero copy which was
// deliberately rewritten to stop duplicating the home page profile.
func DefaultAboutContent() AboutContent {
	return AboutContent{
		AboutKicker:              "About the keeper",
		AboutHeading:             "关于我，也关于这座博客",
		AboutLead:                "我是这座博客的维护者，也是一名长期主义的记录者。比起追逐热点，更在意那些经得起时间检验的工程实践与真实判断。",
		AboutSignature:           "记录与维护，皆在字里行间",
		AboutContactLabel:        "和我联系",
		AboutReadingLabel:        "阅读文章",
		AboutPrinciplesKicker:    "Publishing principles",
		AboutPrinciplesHeading:   "让内容按自己的节奏生长",
		AboutPrinciplesIntro:     "这些原则约束这个博客的设计与维护方式，也帮助阅读始终停留在内容本身。",
		Principle1Title:          "内容先于装饰",
		Principle1Description:    "排版、留白与动效都服务于理解；移除背景效果之后，文章仍然应该清楚而完整。",
		Principle2Title:          "让系统承担复杂",
		Principle2Description:    "主题、组件和状态保持稳定契约，把维护成本留在系统内部，而不是交给每一篇内容。",
		Principle3Title:          "为长期阅读留白",
		Principle3Description:    "不追逐每一次短暂变化，让归档、链接与文字在更长时间里仍然可以被重新找到。",
		AboutContactKicker:       "Say hello",
		AboutContactHeadingWith:  "在这些地方找到我",
		AboutContactHeadingEmpty: "联系方式暂时停泊",
		AboutContactIntroWith:    "选择你习惯的平台，继续聊写作、技术或长期维护。",
		AboutContactIntroEmpty:   "站点尚未公开社交链接，你仍可以从归档继续阅读。",
		AboutContactEmptyCta:     "先读一篇文章",
		AboutPortraitLine1:       "Blog keeper",
		AboutPortraitLine2:       "记录，是为了更好地想起",
	}
}

// IsValidAboutContent reports whether every field satisfies its rune cap.
// Empty values are valid so NormalizeAboutContent can reset them to the
// shipped defaults.
func IsValidAboutContent(value AboutContent) bool {
	return !exceedsRuneLimit(value.AboutKicker, AboutKickerMaxRunes) &&
		!exceedsRuneLimit(value.AboutHeading, AboutHeadingMaxRunes) &&
		!exceedsRuneLimit(value.AboutLead, AboutLeadMaxRunes) &&
		!exceedsRuneLimit(value.AboutSignature, AboutSignatureMaxRunes) &&
		!exceedsRuneLimit(value.AboutContactLabel, AboutActionLabelMaxRunes) &&
		!exceedsRuneLimit(value.AboutReadingLabel, AboutActionLabelMaxRunes) &&
		!exceedsRuneLimit(value.AboutPrinciplesKicker, AboutPrinciplesKickerMaxRunes) &&
		!exceedsRuneLimit(value.AboutPrinciplesHeading, AboutPrinciplesHeadingMaxRunes) &&
		!exceedsRuneLimit(value.AboutPrinciplesIntro, AboutPrinciplesIntroMaxRunes) &&
		!exceedsRuneLimit(value.Principle1Title, AboutPrincipleTitleMaxRunes) &&
		!exceedsRuneLimit(value.Principle1Description, AboutPrincipleDescriptionMaxRunes) &&
		!exceedsRuneLimit(value.Principle2Title, AboutPrincipleTitleMaxRunes) &&
		!exceedsRuneLimit(value.Principle2Description, AboutPrincipleDescriptionMaxRunes) &&
		!exceedsRuneLimit(value.Principle3Title, AboutPrincipleTitleMaxRunes) &&
		!exceedsRuneLimit(value.Principle3Description, AboutPrincipleDescriptionMaxRunes) &&
		!exceedsRuneLimit(value.AboutContactKicker, AboutContactKickerMaxRunes) &&
		!exceedsRuneLimit(value.AboutContactHeadingWith, AboutContactHeadingMaxRunes) &&
		!exceedsRuneLimit(value.AboutContactHeadingEmpty, AboutContactHeadingMaxRunes) &&
		!exceedsRuneLimit(value.AboutContactIntroWith, AboutContactIntroMaxRunes) &&
		!exceedsRuneLimit(value.AboutContactIntroEmpty, AboutContactIntroMaxRunes) &&
		!exceedsRuneLimit(value.AboutContactEmptyCta, AboutContactEmptyCtaMaxRunes) &&
		!exceedsRuneLimit(value.AboutPortraitLine1, AboutPortraitLine1MaxRunes) &&
		!exceedsRuneLimit(value.AboutPortraitLine2, AboutPortraitLine2MaxRunes)
}

// NormalizeAboutContent returns a complete AboutContent by trimming every
// field and falling back to the corresponding default when the input is
// empty or longer than the per-field rune cap.
func NormalizeAboutContent(value AboutContent) AboutContent {
	defaults := DefaultAboutContent()
	return AboutContent{
		AboutKicker:              normalizeSectionText(value.AboutKicker, defaults.AboutKicker, AboutKickerMaxRunes),
		AboutHeading:             normalizeSectionText(value.AboutHeading, defaults.AboutHeading, AboutHeadingMaxRunes),
		AboutLead:                normalizeSectionText(value.AboutLead, defaults.AboutLead, AboutLeadMaxRunes),
		AboutSignature:           normalizeSectionText(value.AboutSignature, defaults.AboutSignature, AboutSignatureMaxRunes),
		AboutContactLabel:        normalizeSectionText(value.AboutContactLabel, defaults.AboutContactLabel, AboutActionLabelMaxRunes),
		AboutReadingLabel:        normalizeSectionText(value.AboutReadingLabel, defaults.AboutReadingLabel, AboutActionLabelMaxRunes),
		AboutPrinciplesKicker:    normalizeSectionText(value.AboutPrinciplesKicker, defaults.AboutPrinciplesKicker, AboutPrinciplesKickerMaxRunes),
		AboutPrinciplesHeading:   normalizeSectionText(value.AboutPrinciplesHeading, defaults.AboutPrinciplesHeading, AboutPrinciplesHeadingMaxRunes),
		AboutPrinciplesIntro:     normalizeSectionText(value.AboutPrinciplesIntro, defaults.AboutPrinciplesIntro, AboutPrinciplesIntroMaxRunes),
		Principle1Title:          normalizeSectionText(value.Principle1Title, defaults.Principle1Title, AboutPrincipleTitleMaxRunes),
		Principle1Description:    normalizeSectionText(value.Principle1Description, defaults.Principle1Description, AboutPrincipleDescriptionMaxRunes),
		Principle2Title:          normalizeSectionText(value.Principle2Title, defaults.Principle2Title, AboutPrincipleTitleMaxRunes),
		Principle2Description:    normalizeSectionText(value.Principle2Description, defaults.Principle2Description, AboutPrincipleDescriptionMaxRunes),
		Principle3Title:          normalizeSectionText(value.Principle3Title, defaults.Principle3Title, AboutPrincipleTitleMaxRunes),
		Principle3Description:    normalizeSectionText(value.Principle3Description, defaults.Principle3Description, AboutPrincipleDescriptionMaxRunes),
		AboutContactKicker:       normalizeSectionText(value.AboutContactKicker, defaults.AboutContactKicker, AboutContactKickerMaxRunes),
		AboutContactHeadingWith:  normalizeSectionText(value.AboutContactHeadingWith, defaults.AboutContactHeadingWith, AboutContactHeadingMaxRunes),
		AboutContactHeadingEmpty: normalizeSectionText(value.AboutContactHeadingEmpty, defaults.AboutContactHeadingEmpty, AboutContactHeadingMaxRunes),
		AboutContactIntroWith:    normalizeSectionText(value.AboutContactIntroWith, defaults.AboutContactIntroWith, AboutContactIntroMaxRunes),
		AboutContactIntroEmpty:   normalizeSectionText(value.AboutContactIntroEmpty, defaults.AboutContactIntroEmpty, AboutContactIntroMaxRunes),
		AboutContactEmptyCta:     normalizeSectionText(value.AboutContactEmptyCta, defaults.AboutContactEmptyCta, AboutContactEmptyCtaMaxRunes),
		AboutPortraitLine1:       normalizeSectionText(value.AboutPortraitLine1, defaults.AboutPortraitLine1, AboutPortraitLine1MaxRunes),
		AboutPortraitLine2:       normalizeSectionText(value.AboutPortraitLine2, defaults.AboutPortraitLine2, AboutPortraitLine2MaxRunes),
	}
}

// CloneAboutContent returns a value-level copy. All fields are strings so a
// shallow struct copy is sufficient.
func CloneAboutContent(value AboutContent) AboutContent {
	return value
}
