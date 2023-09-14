package topology

import (
	"fmt"
	"math/rand"
)

type Generator interface {
	GenerateKey() string
	GenerateValue() string
}

type RandomTagGenerator struct {
	random *rand.Rand
}

func (rtg *RandomTagGenerator) GenerateKey() string {
	return fmt.Sprintf("%v-%v", rtg.randomString(ADJECTIVES), rtg.randomString(NOUNS))
}

func (rtg *RandomTagGenerator) GenerateValue() string {
	return fmt.Sprintf("%v-%v", rtg.randomString(ADVERBS), rtg.randomString(VERBS))
}

func (rtg *RandomTagGenerator) randomString(words []string) string {
	return words[rtg.random.Intn(len(words))]
}

var (
	// ADJECTIVES ...
	ADJECTIVES = []string{"autumn", "hidden", "bitter", "misty", "silent", "empty", "dry", "dark", "summer",
		"icy", "delicate", "quiet", "white", "cool", "spring", "winter", "patient",
		"twilight", "dawn", "crimson", "wispy", "weathered", "blue", "billowing",
		"broken", "cold", "damp", "falling", "frosty", "green", "long", "late", "lingering",
		"bold", "little", "morning", "muddy", "old", "red", "rough", "still", "small",
		"sparkling", "throbbing", "shy", "wandering", "withered", "wild", "black",
		"young", "holy", "solitary", "fragrant", "aged", "snowy", "proud", "floral",
		"restless", "divine", "polished", "ancient", "purple", "lively", "nameless"}

	// NOUNS ...
	NOUNS = []string{"waterfall", "river", "breeze", "moon", "rain", "wind", "sea", "morning",
		"snow", "lake", "sunset", "pine", "shadow", "leaf", "dawn", "glitter", "forest",
		"hill", "cloud", "meadow", "sun", "glade", "bird", "brook", "butterfly",
		"bush", "dew", "dust", "field", "fire", "flower", "firefly", "feather", "grass",
		"haze", "mountain", "night", "pond", "darkness", "snowflake", "silence",
		"sound", "sky", "shape", "surf", "thunder", "violet", "water", "wildflower",
		"wave", "water", "resonance", "sun", "wood", "dream", "cherry", "tree", "fog",
		"frost", "voice", "paper", "frog", "smoke", "star"}

	VERBS = []string{"run","jump","sing","dance","write","read","swim","paint","cook","eat","sleep",
		"drive","talk","listen","think","play","study","work","laugh","cry","climb","fly","build","plant",
		"create","design","calculate","solve","program","code","investigate","explore","imagine","craft",
		"sew","draw","sculpt","sing","type","analyze","communicate","collaborate","travel","relax",
		"meditate","exercise","hike","capture","invent","inspire"}

  ADVERBS = []string{"quickly","slowly","quietly","loudly","carefully","boldly","cautiously","anxiously",
	  "happily","sadly","bravely","eagerly","gently","roughly","softly","smoothly","suddenly","gradually",
	  "frequently","rarely","always","never","shortly","longingly","mysteriously","calmly","nervously","honestly",
	  "deftly","warily","warmly","coolly","effortlessly","clumsily","politely","impolitely","seriously",
	  "playfully","steadily","vigorously","sloppily","painstakingly","regularly","unusually","seldom","eventually",
	  "readily","reluctantly","accidentally","intentionally","quickly"}
)