package pl

// kinshipVariantEntries keeps Polish gender-specific kinship nouns as distinct
// lexemes rather than inventing a single gender-neutral form.
var kinshipVariantEntries = []lexicalEntry{
	{"pl.ojczym.n", "ojczym", "noun", []lexicalMeaning{{"pl.ojczym.stepparent", "A male person recognized as a parent through partnership with a child's parent.", "stepparent"}}},
	{"pl.macocha.n", "macocha", "noun", []lexicalMeaning{{"pl.macocha.stepparent", "A female person recognized as a parent through partnership with a child's parent.", "stepparent"}}},
	{"pl.dziadek.n", "dziadek", "noun", []lexicalMeaning{{"pl.dziadek.grandparent", "A male parent of a person's parent, including recognized social kinship.", "grandparent"}}},
	{"pl.babcia.n", "babcia", "noun", []lexicalMeaning{{"pl.babcia.grandparent", "A female parent of a person's parent, including recognized social kinship.", "grandparent"}}},
	{"pl.wnuk.n", "wnuk", "noun", []lexicalMeaning{{"pl.wnuk.grandchild", "A male child of a person's child, including recognized social kinship.", "grandchild"}}},
	{"pl.wnuczka.n", "wnuczka", "noun", []lexicalMeaning{{"pl.wnuczka.grandchild", "A female child of a person's child, including recognized social kinship.", "grandchild"}}},
	{"pl.wujek.n", "wujek", "noun", []lexicalMeaning{{"pl.wujek.aunt_uncle", "A male sibling of a parent or equivalent relative recognized in that kinship role.", "aunt_uncle"}}},
	{"pl.ciotka.n", "ciotka", "noun", []lexicalMeaning{{"pl.ciotka.aunt_uncle", "A female sibling of a parent or equivalent relative recognized in that kinship role.", "aunt_uncle"}}},
	{"pl.kuzyn.n", "kuzyn", "noun", []lexicalMeaning{{"pl.kuzyn.cousin", "A male person related through collateral kinship rather than direct descent or siblinghood.", "cousin"}}},
	{"pl.kuzynka.n", "kuzynka", "noun", []lexicalMeaning{{"pl.kuzynka.cousin", "A female person related through collateral kinship rather than direct descent or siblinghood.", "cousin"}}},
}
