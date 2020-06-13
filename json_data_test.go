package hapijson

var jsonInvalidTestSet = []string{
	// corrupted json
	// test string
	`string value`,
	`"string value`,
	`string value"`,

	// unicode and hex
	`"\uabaz"`,
	`"\uab"`,
	`"\ua"`,
	`"\u"`,
	`"\x.3"`,
	`"\x1"`,

	`"\"`,

	// test numbers
	`e32432`,
	`01.234`,
	`0e.32432`,
	`0.324.32`,
	`1234e.56789`,
	`+123456.e-789`,
	`-123456.-e789`,
	`-123456.+E789`,
	`-123456.789e`,
	`-123456.e7e89`,

	// test boolean and null
	`tru`,
	" truE",
	" true,",
	`fal`,
	`falSe `,
	` true 1`,
	`nul`,
	` nulL `,
	` nu\lL `,

	// test object and array
	`[`,
	`[[]`,
	`{`,
	`{{}`,
	`[{]}`,
	`[{"num":1234]}`,
	`[{},{},3k]`,

	`{"k":"val"`,
	`43:"val"}`,
	`{43:"val"}`,
	`{:"val"}`,
	`{key:val}`,
	`{"key:val}`,
	`{"key"val}`,
	`{"key" val}`,
	`{"key"\: "val"}`,
	`{"key"::val}`,
	`{"key":val}`,
	`{"key":"val}`,
	`{"key":"val",}`,
	`{"key":{"k2"}}`,
	`{"key":"val",, "k2": 234}`,
	`{"key":"val", "k2": 234, [}`,
	`{"key":"val", "k2": 234 }{`,
	`{}[]`,

	`[{]`,
	`["v1" "v2"]`,
	`["v1":"v2",]`,
	`[{},]`,
	`[{"k": 1,}]`,
	`[{},value]`,
	`[{},"fd",, 34]`,
	`[{},"fd", 34],`,
	`[{\},"fd", 34],`,
	`[{},"fd", 34]{}`,
}

var jsonValidTestSet = []string{
	// string
	`"string value"`,
	`"string \"value\""`,

	// unicode and hex
	`"\u0032"`,
	`"\ud83d\ude02"`,
	`"\ud83d"`,
	`"\x32"`,
	`"\xe8"`,

	// numbers
	`1234`,
	`.1234`,
	`1234.`,
	`0.1234`,
	`0.e1234`,
	`0.e-1234`,
	`0.e+1234`,
	`0.E-1234`,
	`0.E+1234`,
	`0.1e-234`,
	`0.1E-234`,
	`0.1e+234`,
	`0.1E+234`,

	// boolean and null
	`false`,
	`true`,
	`null`,

	// object
	`{}`,
	`{"key": "234"}`,
	`{"key": 234}`,
	`{"key": true}`,
	`{"key": false}`,
	`{"key": false, "key2": "val2", "key3": []}`,

	// array
	`[]`,
	`[1, 2, 3,4]`,
	`["1", "2", 3,4]`,
	`["1", "2", 3,4, false, true, null, {}]`,
	`["1", "2", 3,4, false, true, null, {"key": "val"}]`,

	// whites test
	`[       ]`, `{    "k"     :    "34",  "a r r ay"  : [ 43 , 4 , "1" ,  false , true  ,     
        null ]}`,
	`         {           }       `,
	`{ "232" 
        : 
        "val 
        enter"
        ,
        "

                the key":
        [
        "
        so what?
        "

                ,
    2
        ]
    }`,
}

var jsonFromGoogleTranslator = `
[
  [
    ["一个", "a", null, null, 1],
    [null, null, "Yīgè", "ā,ə"]
  ],
  [
    [
      "冠词",
      ["一个", "一"],
      [
        ["一个", ["a", "an"], null, 0.121313766],
        ["一", ["a"], null, 0.036998607]
      ],
      "a",
      13
    ]
  ],
  "en",
  null,
  null,
  [
    [
      "a",
      null,
      [
        ["一个", 1000, true, false],
        ["一种", 1000, true, false],
        ["A", 1000, true, false],
        ["一", 0, true, false]
      ],
      [[0, 1]],
      "a",
      0,
      0
    ]
  ],
  0.0,
  [],
  [["en"], null, [0.0], ["en"]],
  null,
  null,
  [
    [
      "名词",
      [
        [["angstrom", "angstrom unit"], ""],
        [["amp", "ampere"], ""],
        [["axerophthol"], ""],
        [["adenine"], ""]
      ],
      "a"
    ]
  ],
  [
    [
      "冠词",
      [
        [
          "used when referring to someone or something for the first time in a text or conversation.",
          "m_en_us1219181.001",
          "a man came out of the room"
        ],
        [
          "used to indicate membership of a class of people or things.",
          "m_en_us1219181.006",
          "She's a banker, married to a stockbroker, and they have a daughter about the same age as Amy."
        ],
        [
          "used when expressing rates or ratios; in, to, or for each; per.",
          "m_en_us1219181.007",
          "You can't drive over five miles an hour down any street in New York."
        ]
      ],
      "a"
    ]
  ],
  [
    [
      [
        "Regarding academic medicine, it has become increasingly difficult for \u003cb\u003ea\u003c/b\u003e Freud or a Mendel to gain recognition without university affiliation or corporate sponsorship.",
        null,
        null,
        null,
        3,
        "m_en_us1219181.005"
      ],
      [
        "Bob's conducting a three-year internet romance with \u003cb\u003ea\u003c/b\u003e girl he's never met.",
        null,
        null,
        null,
        3,
        "m_en_us1219181.001"
      ],
      [
        "About \u003cb\u003ea\u003c/b\u003e mile further down the road, another dog ran out in front of the taxi.",
        null,
        null,
        null,
        3,
        "m_en_us1219181.002"
      ],
      [
        "The latest letter was from \u003cb\u003ea\u003c/b\u003e Mrs Singh, who complained about two radio stations.",
        null,
        null,
        null,
        3,
        "m_en_us1219181.004"
      ],
      [
        "Called \u003cb\u003ea\u003c/b\u003e Judas by his countrymen, he received an elbow from another player, and left the pitch injured.",
        null,
        null,
        null,
        3,
        "m_en_us1219181.005"
      ],
      [
        "Lilly is \u003cb\u003ea\u003c/b\u003e Siamese cat who survived a two-week cross-country move while stuck in a drawer.",
        null,
        null,
        null,
        3,
        "m_en_us1219181.006"
      ],
      [
        "Does anyone know \u003cb\u003ea\u003c/b\u003e Mr Daeller?",
        null,
        null,
        null,
        3,
        "m_en_us1219181.004"
      ],
      [
        "She was born in about 1670, the daughter of \u003cb\u003ea\u003c/b\u003e Mr Freeman of Holbeach in Lincolnshire.",
        null,
        null,
        null,
        3,
        "m_en_us1219181.004"
      ],
      [
        "Children need \u003cb\u003ea\u003c/b\u003e place for their computer equipment, and parents need closet space for their clothing.",
        null,
        null,
        null,
        3,
        "m_en_us1219181.001"
      ],
      [
        "typing 60 words \u003cb\u003ea\u003c/b\u003e minute",
        null,
        null,
        null,
        3,
        "m_en_gb0000030.007"
      ],
      [
        "My mom's \u003cb\u003ea\u003c/b\u003e pharmacist and my dad's a realtor.",
        null,
        null,
        null,
        3,
        "m_en_us1219181.006"
      ],
      [
        "The film looks fantastic: there is not \u003cb\u003ea\u003c/b\u003e spot, or a scratch, or a visual defect to be seen.",
        null,
        null,
        null,
        3,
        "m_en_us1219181.003"
      ],
      [
        "We had to write \u003cb\u003ea\u003c/b\u003e story about a natural disaster for creative writing.",
        null,
        null,
        null,
        3,
        "m_en_us1219181.001"
      ],
      [
        "On September 29 a letter arrived at our address for \u003cb\u003ea\u003c/b\u003e Ms L Doherty.",
        null,
        null,
        null,
        3,
        "m_en_us1219181.004"
      ],
      [
        "I had to own up to the fact that I'd never read \u003cb\u003ea\u003c/b\u003e word by Crofts.",
        null,
        null,
        null,
        3,
        "m_en_us1219181.003"
      ],
      [
        "I type 15 words \u003cb\u003ea\u003c/b\u003e minute with a lot of mistakes.",
        null,
        null,
        null,
        3,
        "m_en_us1219181.007"
      ],
      [
        "I look at these miserable people, and wouldn't trade my life with theirs for \u003cb\u003ea\u003c/b\u003e million dollars.",
        null,
        null,
        null,
        3,
        "m_en_us1219181.002"
      ],
      [
        "I stopped to pick up \u003cb\u003ea\u003c/b\u003e gallon of milk on my way home from work.",
        null,
        null,
        null,
        3,
        "m_en_us1219181.002"
      ],
      [
        "Incensed at the fiasco, I went back to the website to try and find a telephone number to call - not \u003cb\u003ea\u003c/b\u003e thing!",
        null,
        null,
        null,
        3,
        "m_en_us1219181.003"
      ],
      [
        "Most refugees say they never saw \u003cb\u003ea\u003c/b\u003e drop of food aid - despite almost one million tonnes flooding into the country every year.",
        null,
        null,
        null,
        3,
        "m_en_us1219181.003"
      ],
      [
        "Notice that every car seen in the show is \u003cb\u003ea\u003c/b\u003e Chevrolet, out of consideration for their sponsor.",
        null,
        null,
        null,
        3,
        "m_en_us1219181.006"
      ],
      [
        "\u003cb\u003ea\u003c/b\u003e Mr Smith telephoned",
        null,
        null,
        null,
        3,
        "m_en_gb0000030.004"
      ],
      [
        "In 1984 he was granted his fervent wish to acquire \u003cb\u003ea\u003c/b\u003e Picasso.",
        null,
        null,
        null,
        3,
        "m_en_us1219181.006"
      ],
      [
        "I think there's not \u003cb\u003ea\u003c/b\u003e person born that doesn't have a gift to offer in some way.",
        null,
        null,
        null,
        3,
        "m_en_us1219181.003"
      ],
      [
        "An internal report written by \u003cb\u003ea\u003c/b\u003e manager at the nuclear waste reprocessing plant was leaked this week.",
        null,
        null,
        null,
        3,
        "m_en_us1219181.001"
      ],
      [
        "Jack crouched down and hid behind \u003cb\u003ea\u003c/b\u003e tree trunk.",
        null,
        null,
        null,
        3,
        "m_en_us1219181.001"
      ],
      [
        "\u003cb\u003ea\u003c/b\u003e Mr. Smith telephoned",
        null,
        null,
        null,
        3,
        "m_en_us1219181.004"
      ],
      [
        "\u003cb\u003ea\u003c/b\u003e ringing of bells",
        null,
        null,
        null,
        3,
        "neid_11"
      ],
      [
        "The truckers are angry at the rise in diesel prices, which currently average 81.3p \u003cb\u003ea\u003c/b\u003e litre.",
        null,
        null,
        null,
        3,
        "m_en_us1219181.007"
      ],
      [
        "15 euro \u003cb\u003ea\u003c/b\u003e metre",
        null,
        null,
        null,
        3,
        "neid_7"
      ]
    ]
  ]
]

`
