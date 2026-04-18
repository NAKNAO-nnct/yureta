package model

// BasicData は全メッセージ共通のフィールドです。
type BasicData struct {
	ID   string `json:"id"`
	Code int    `json:"code"`
	Time string `json:"time"`
}

// --- 地震情報 (code: 551) ---

// JMAQuake は気象庁の地震情報です。
type JMAQuake struct {
	BasicData
	Issue      JMAQuakeIssue `json:"issue"`
	Earthquake *Earthquake   `json:"earthquake"`
	Points     []Point       `json:"points"`
}

type JMAQuakeIssue struct {
	Source  string `json:"source"`
	Time    string `json:"time"`
	Type    string `json:"type"`
	Correct string `json:"correct"`
}

type Earthquake struct {
	Time            string     `json:"time"`
	Hypocenter      Hypocenter `json:"hypocenter"`
	MaxScale        int        `json:"maxScale"`
	DomesticTsunami string     `json:"domesticTsunami"`
	ForeignTsunami  string     `json:"foreignTsunami"`
}

type Hypocenter struct {
	Name      string  `json:"name"`
	ReduceName string `json:"reduceName"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Depth     int     `json:"depth"`
	Magnitude float64 `json:"magnitude"`
}

type Point struct {
	Pref  string `json:"pref"`
	Addr  string `json:"addr"`
	IsArea bool   `json:"isArea"`
	Scale int    `json:"scale"`
}

// --- 津波予報 (code: 552) ---

// JMATsunami は気象庁の津波予報です。
type JMATsunami struct {
	BasicData
	Cancelled bool             `json:"cancelled"`
	Issue     JMATsunamiIssue  `json:"issue"`
	Areas     []TsunamiArea    `json:"areas"`
}

type JMATsunamiIssue struct {
	Source string `json:"source"`
	Time   string `json:"time"`
	Type   string `json:"type"`
}

type TsunamiArea struct {
	Grade     string          `json:"grade"`
	Immediate bool            `json:"immediate"`
	Name      string          `json:"name"`
	FirstHeight *FirstHeight  `json:"firstHeight"`
	MaxHeight   *MaxHeight    `json:"maxHeight"`
}

type FirstHeight struct {
	ArrivalTime string `json:"arrivalTime"`
	Condition   string `json:"condition"`
}

type MaxHeight struct {
	Description string  `json:"description"`
	Value       float64 `json:"value"`
}

// --- 緊急地震速報 発表検出 (code: 554) ---

// EEWDetection は緊急地震速報の発表検出です。
type EEWDetection struct {
	BasicData
	Type string `json:"type"`
}

// --- 緊急地震速報・警報 (code: 556) ---

// EEW は緊急地震速報（警報）です。
type EEW struct {
	BasicData
	Test       bool         `json:"test"`
	Cancelled  bool         `json:"cancelled"`
	Issue      EEWIssue     `json:"issue"`
	Earthquake *EEWQuake    `json:"earthquake"`
	Areas      []EEWArea    `json:"areas"`
}

type EEWIssue struct {
	Time    string `json:"time"`
	EventID string `json:"eventId"`
	Serial  string `json:"serial"`
}

type EEWQuake struct {
	OriginTime  string      `json:"originTime"`
	ArrivalTime string      `json:"arrivalTime"`
	Condition   string      `json:"condition"`
	Hypocenter  Hypocenter  `json:"hypocenter"`
}

type EEWArea struct {
	Pref        string  `json:"pref"`
	Name        string  `json:"name"`
	ScaleFrom   float64 `json:"scaleFrom"`
	ScaleTo     float64 `json:"scaleTo"`
	KindCode    string  `json:"kindCode"`
	ArrivalTime *string `json:"arrivalTime"`
}

// ScaleLabel は震度コードを日本語表記に変換します。
func ScaleLabel(scale int) string {
	switch scale {
	case 10:
		return "震度1"
	case 20:
		return "震度2"
	case 30:
		return "震度3"
	case 40:
		return "震度4"
	case 45:
		return "震度5弱"
	case 50:
		return "震度5強"
	case 55:
		return "震度6弱"
	case 60:
		return "震度6強"
	case 70:
		return "震度7"
	default:
		return "不明"
	}
}
