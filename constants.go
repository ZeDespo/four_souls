package four_souls

// Rule constants
const (
	playLootCard        uint8 = 0
	buyItem             uint8 = 1
	attackMonster       uint8 = 2
	activateCharacter   uint8 = 3
	activateItem        uint8 = 4
	endActivePlayerTurn uint8 = 5
	doNothing           uint8 = 6
	peekTheresOptions   uint8 = 7
	forceAttackDeck     int8  = -1
	forceAttackMon      int8  = -2
)

// !!! ID NUMBERS FOR THE CARDS!!! \\

// LOOT CARDS

// basic loot
const (
	aPenny           uint16 = 1
	twoCents         uint16 = 2
	threeCents       uint16 = 3
	fourCents        uint16 = 4
	aNickel          uint16 = 5
	aDime            uint16 = 6
	blankRune        uint16 = 7
	bomb             uint16 = 8
	butterBean       uint16 = 9
	dagaz            uint16 = 10
	diceShard        uint16 = 11
	ehwaz            uint16 = 12
	goldBomb         uint16 = 13
	lilBattery       uint16 = 14
	megaBattery      uint16 = 15
	pillsBlue        uint16 = 16
	pillsRed         uint16 = 17
	pillsYellow      uint16 = 18
	soulHeart        uint16 = 19
	aSack            uint16 = 20
	chargedPenny     uint16 = 21
	creditCard       uint16 = 22
	holyCard         uint16 = 23
	jera             uint16 = 24
	joker            uint16 = 25
	pillsPurple      uint16 = 26
	ansuz            uint16 = 27
	blackRune        uint16 = 28
	getOutOfJail     uint16 = 29
	goldKey          uint16 = 30
	perthro          uint16 = 31
	pillsBlack       uint16 = 32
	pillsSpots       uint16 = 33
	pillsWhite       uint16 = 34
	questionMarkCard uint16 = 35
	lostSoul         uint16 = 36
	twoOfDiamonds    uint16 = 37
)

// Trinkets
const (
	bloodyPenny      uint16 = 40
	brokenAnkh       uint16 = 41
	cainsEye         uint16 = 42
	counterfeitPenny uint16 = 43
	curvedHorn       uint16 = 44
	goldenHorseShoe  uint16 = 45
	guppysHairball   uint16 = 46
	purpleHeart      uint16 = 47
	swallowedPenny   uint16 = 48
	cancer           uint16 = 49
	pinkEye          uint16 = 50
	aaaBattery       uint16 = 51
	pokerChip        uint16 = 52
	tapeWorm         uint16 = 53
	theLeftHand      uint16 = 54
)

// Tarot cards
const (
	theFool          uint16 = 60
	theMagician      uint16 = 61
	theHighPriestess uint16 = 62
	theEmpress       uint16 = 63
	theEmperor       uint16 = 64
	theHierophant    uint16 = 65
	theLovers        uint16 = 66
	theChariot       uint16 = 67
	justice          uint16 = 68
	theHermit        uint16 = 69
	wheelOfFortune   uint16 = 70
	strength         uint16 = 71
	theHangedMan     uint16 = 72
	deathLoot        uint16 = 73
	theTower         uint16 = 74
	theDevil         uint16 = 75
	temperance       uint16 = 76
	theStars         uint16 = 77
	theMoon          uint16 = 78
	theSun           uint16 = 79
	judgement        uint16 = 80
	theWorld         uint16 = 81
)

// END LOOT CARDS

// MONSTER CARDS
// Note: If the monster does not have an cardEffect attached to it, they will have an id of 0 and will not be included here.

// Basic Monsters
const (
	bigSpider    uint16 = 90
	blackBony    uint16 = 91
	boomFly      uint16 = 92
	dankGlobin   uint16 = 93
	dinga        uint16 = 94
	dople        uint16 = 95
	evilTwin     uint16 = 96
	greedling    uint16 = 97
	hanger       uint16 = 98
	hopper       uint16 = 99
	horf         uint16 = 100
	keeperHead   uint16 = 101
	leaper       uint16 = 102
	momsDeadHand uint16 = 103
	momsEye      uint16 = 104
	momsHand     uint16 = 105
	mulliboom    uint16 = 106
	mulligan     uint16 = 107
	portal       uint16 = 108
	psyHorf      uint16 = 109
	rageCreep    uint16 = 110
	ringOfFlies  uint16 = 111
	stoney       uint16 = 112
	swarmOfFlies uint16 = 113
	wizoob       uint16 = 114
	begotten     uint16 = 115
	boil         uint16 = 116
	deathsHead   uint16 = 117
	gaper        uint16 = 118
	imp          uint16 = 119
	knight       uint16 = 120
	parabite     uint16 = 121
	ragling      uint16 = 122
	roundWorm    uint16 = 123
	bony         uint16 = 124
	brain        uint16 = 125
	flaminHopper uint16 = 126
	globin       uint16 = 127
	roundy       uint16 = 128
	sucker       uint16 = 129
	swarmer      uint16 = 130
	tumor        uint16 = 131
)

// monsters, bosses, mega bosses, with no effect
const (
	clotty         = 500
	codWorm        = 501
	conjoinedFatty = 502
	dip            = 503
	fatBat         = 504
	fatty          = 505
	fly            = 506
	leech          = 507
	paleFatty      = 508
	pooter         = 509
	redHost        = 510
	spider         = 511
	squirt         = 512
	trite          = 513
	charger        = 514
	nerveEnding    = 515
	gurdy          = 516
	littleHorn     = 517
	monstro        = 518
	theCage        = 519
	widow          = 520
	hush           = 521
)

// cursed / holy monsters
const (
	cursedFatty      uint16 = 140
	cursedGaper      uint16 = 141
	cursedHorf       uint16 = 142
	cursedKeeperHead uint16 = 143
	cursedMomsHand   uint16 = 144
	cursedPsyHorf    uint16 = 145
	holyDinga        uint16 = 146
	holyDip          uint16 = 147
	holyKeeperHead   uint16 = 148
	holyMomsEye      uint16 = 149
	holySquirt       uint16 = 150
	cursedGlobin     uint16 = 151
	cursedTumor      uint16 = 152
	holyBony         uint16 = 153
	holyMulligan     uint16 = 154
)

// bosses
const (
	carrionQueen         uint16 = 160
	chub                 uint16 = 162
	conquest             uint16 = 163
	daddyLongLegsMonster uint16 = 164
	darkOne              uint16 = 165
	deathMonster         uint16 = 166
	delirium             uint16 = 167
	envy                 uint16 = 168
	famine               uint16 = 169
	gemini               uint16 = 170
	gluttony             uint16 = 171
	greedMonster         uint16 = 172
	gurdyJr              uint16 = 173
	larryJr              uint16 = 174
	lust                 uint16 = 175
	maskOfInfamy         uint16 = 176
	megaFatty            uint16 = 177
	peep                 uint16 = 178
	pestilence           uint16 = 179
	pin                  uint16 = 180
	pride                uint16 = 181
	ragman               uint16 = 182
	scolex               uint16 = 183
	sloth                uint16 = 184
	theBloat             uint16 = 185
	theDukeOfFlies       uint16 = 186
	theHaunt             uint16 = 187
	war                  uint16 = 188
	wrath                uint16 = 189
	fistula              uint16 = 190
	gurglings            uint16 = 191
	polycephalus         uint16 = 192
	steven               uint16 = 193
	blastocyst           uint16 = 194
	dingle               uint16 = 195
	headlessHorseman     uint16 = 196
	krampus              uint16 = 197
	monstroII            uint16 = 198
	theFallen            uint16 = 199
)

// Mega bosses
const (
	mom          uint16 = 200
	satan        uint16 = 201
	theLamb      uint16 = 202
	isaacMonster uint16 = 203
	momsHeart    uint16 = 204
)

// random happenings
const (
	ambush           uint16 = 210
	chest            uint16 = 211
	cursedChest      uint16 = 212
	darkChest        uint16 = 213
	devilDeal        uint16 = 214
	goldChest        uint16 = 215
	greedHappening   uint16 = 216
	iCanSeeForever   uint16 = 217
	trollBombs       uint16 = 218
	megaTrollBomb    uint16 = 219
	secretRoom       uint16 = 220
	shopUpgrade      uint16 = 221
	weNeedToGoDeeper uint16 = 222
	xlFloor          uint16 = 223
	iAmError         uint16 = 224
	trapDoor         uint16 = 225
	angelRoom        uint16 = 226
	bossRush         uint16 = 227
	headTrauma       uint16 = 228
	holyChest        uint16 = 229
	spikedChest      uint16 = 230
)

// curses
const (
	curseOfAmnesia   uint16 = 240
	curseOfGreed     uint16 = 241
	curseOfLoss      uint16 = 242
	curseOfPain      uint16 = 243
	curseOfTheBlind  uint16 = 244
	curseOfFatigue   uint16 = 245
	curseOfTinyHands uint16 = 246
	curseOfBloodLust uint16 = 247
	curseOfImpulse   uint16 = 248
)

// END LOOT CARDS

// TREASURE CARDS

// Starting items (passive and active)
const (
	theD6         uint16 = 250
	yumHeart      uint16 = 251
	sleightOfHand uint16 = 252
	bookOfBelial  uint16 = 253
	foreverAlone  uint16 = 254
	theCurse      uint16 = 255
	bloodLust     uint16 = 256
	lazarusRags   uint16 = 257
	incubus       uint16 = 258
	theBone       uint16 = 259
	lordOfThePit  uint16 = 260
	theHolyMantle uint16 = 261
	void          uint16 = 262
	woodenNickel  uint16 = 263
	bagOTrash     uint16 = 264
	darkArts      uint16 = 265
	gimpy         uint16 = 266
	infestation   uint16 = 267
)

// activated items (tap to activate)
const (
	theBattery                   uint16 = 270
	theBible                     uint16 = 271
	blankCard                    uint16 = 272
	bookOfSin                    uint16 = 273
	boomerang                    uint16 = 274
	box                          uint16 = 275
	bumFriend                    uint16 = 276
	compost                      uint16 = 277
	chaos                        uint16 = 278
	chaosCard                    uint16 = 279
	crystalBall                  uint16 = 280
	theD4                        uint16 = 281
	theD20                       uint16 = 282
	theD100                      uint16 = 283
	decoy                        uint16 = 284
	diplopia                     uint16 = 285
	flush                        uint16 = 286
	glassCannon                  uint16 = 287
	godhead                      uint16 = 288
	guppysHead                   uint16 = 289
	guppysPaw                    uint16 = 290
	hostHat                      uint16 = 291
	jawbone                      uint16 = 292
	luckyFoot                    uint16 = 293
	miniMush                     uint16 = 294
	modelingClay                 uint16 = 295
	momsBra                      uint16 = 296
	momsShovel                   uint16 = 297
	monsterManual                uint16 = 298
	mrBoom                       uint16 = 299
	mysterySack                  uint16 = 300
	no                           uint16 = 301
	pandorasBox                  uint16 = 302
	placebo                      uint16 = 303
	potatoPeeler                 uint16 = 304
	razorBlade                   uint16 = 305
	remoteDetonator              uint16 = 306
	sackHead                     uint16 = 307
	sackOfPennies                uint16 = 308
	theShovel                    uint16 = 309
	smartFly                     uint16 = 310
	spoonBender                  uint16 = 311
	twoOfClubs                   uint16 = 312
	crookedPenny                 uint16 = 313
	fruitCake                    uint16 = 314
	iCantBelieveItsNotButterBean uint16 = 315
	lemonMishap                  uint16 = 316
	libraryCard                  uint16 = 317
	ouijaBoard                   uint16 = 318
	planC                        uint16 = 319
	theButterBean                uint16 = 320
	twentyTwenty                 uint16 = 321
	blackCandle                  uint16 = 322
	distantAdmiration            uint16 = 323
	divorcePapers                uint16 = 324
	forgetMeNow                  uint16 = 325
	headOfKrampus                uint16 = 326
	libra                        uint16 = 327
	mutantSpider                 uint16 = 328
	rainbowBaby                  uint16 = 329
	redCandle                    uint16 = 330
)

// paid items (give something to do a thing).
const (
	batteryBum          uint16 = 340
	contractFromBelow   uint16 = 341
	donationMachine     uint16 = 342
	goldenRazorBlade    uint16 = 343
	payToPlay           uint16 = 344
	thePoop             uint16 = 345
	portableSlotMachine uint16 = 346
	smelter             uint16 = 347
	techX               uint16 = 348
	dadsKey             uint16 = 349
	succubus            uint16 = 350
	athame              uint16 = 351
)

// passive items
const (
	babyHaunt             uint16 = 360
	bellyButton           uint16 = 361
	theBlueMap            uint16 = 362
	bobsBrain             uint16 = 363
	breakfast             uint16 = 364
	brimstone             uint16 = 365
	bumbo                 uint16 = 366
	cambionConception     uint16 = 367
	championBelt          uint16 = 368
	chargedBaby           uint16 = 369
	cheeseGrater          uint16 = 370
	theChest              uint16 = 371
	theCompass            uint16 = 372
	curseOfTheTower       uint16 = 373
	theD10                uint16 = 374
	daddyHaunt            uint16 = 375
	dadsLostCoint         uint16 = 376
	darkBum               uint16 = 377
	deadBird              uint16 = 378
	theDeadCat            uint16 = 379
	dinner                uint16 = 380
	dryBaby               uint16 = 381
	edensBlessing         uint16 = 382
	emptyVessel           uint16 = 383
	eyeOfGreed            uint16 = 384
	fannyPack             uint16 = 385
	finger                uint16 = 386
	goatHead              uint16 = 387
	greedsGullet          uint16 = 388
	guppysCollar          uint16 = 389
	theHabit              uint16 = 390
	ipecac                uint16 = 391
	theMap                uint16 = 392
	meat                  uint16 = 393
	theMidasTouch         uint16 = 394
	momsBox               uint16 = 395
	momsCoinPurse         uint16 = 396
	momsPurse             uint16 = 397
	momsRazor             uint16 = 398
	monstrosTooth         uint16 = 399
	thePolaroid           uint16 = 400
	polydactyly           uint16 = 401
	restock               uint16 = 402
	theRelic              uint16 = 403
	sacredHeart           uint16 = 404
	shadow                uint16 = 405
	shinyRock             uint16 = 406
	spiderMod             uint16 = 407
	starterDeck           uint16 = 408
	steamySale            uint16 = 409
	suicideKing           uint16 = 410
	synthoil              uint16 = 411
	tarotCloth            uint16 = 412
	theresOptions         uint16 = 413
	trinityShield         uint16 = 414
	guppysTail            uint16 = 415
	infamy                uint16 = 416
	momsKnife             uint16 = 417
	moreOptions           uint16 = 417
	nineVolt              uint16 = 418
	placenta              uint16 = 419
	skeletonKey           uint16 = 420
	soyMilk               uint16 = 421
	theMissingPage        uint16 = 422
	oneUp                 uint16 = 423
	abaddon               uint16 = 424
	cursedEye             uint16 = 425
	daddyLongLegsTreasure uint16 = 426
	euthanasia            uint16 = 427
	gameBreakingBug       uint16 = 428
	guppysEye             uint16 = 429
	headOfTheKeeper       uint16 = 430
	hourGlass             uint16 = 431
	lard                  uint16 = 432
	magnet                uint16 = 433
	mamaHaunt             uint16 = 434
	momsEyeShaow          uint16 = 435
	phd                   uint16 = 436
	polyphemus            uint16 = 437
	rubberCement          uint16 = 438
	telepathyForDummies   uint16 = 439
	theWiz                uint16 = 440
)

// END TREASURE CARDS

// CHARACTER CARDS

const (
	blueBaby       uint16 = 600
	cain           uint16 = 601
	eden           uint16 = 602
	eve            uint16 = 603
	isaac          uint16 = 604
	judas          uint16 = 605
	lazarus        uint16 = 606
	lilith         uint16 = 607
	maggy          uint16 = 608
	samson         uint16 = 609
	theForgotten   uint16 = 610
	apollyon       uint16 = 611
	azazel         uint16 = 612
	theKeeper      uint16 = 613
	theLost        uint16 = 614
	bumboCharacter uint16 = 615
	darkJudas      uint16 = 616
	guppy          uint16 = 617
	whoreOfBabylon uint16 = 618
)

// !!! END ID CONSTANTS !!! \\

// !!! CARD TEXT !!! \\

//Generic Character card effects.
const characterEffect string = "Play an additional lArea baseCard this turn.\n" +
	"This can be done on any player's turn in response to any action."
const characterEdenEffect string = "Play an additional lArea baseCard this turn.\n" +
	"When you start the game, look at the top 3 cards of the tArea deck." +
	"Choose one, it becomes your Starting Item and gains eternal."

// Starting Items
const foreverAloneDesc = "Choose One:\n" +
	"- Steal 1¢ from a player.\n" +
	"- discardPile a lArea baseCard, then draw a lArea baseCard.\n" +
	"When you take decreaseHP, recharge this."
const sleightOfHandDesc = "Look at the top 3 cards of the deck. Put them back in any order."
const theCurseDesc = "Put the top value of any discardPile pile on top of its deck."
const theD6Desc = "Force a player to re-roll any Dice roll."
const bookOfBelialDesc = "Add or subtract 1 to any Dice roll."
const lazarusRagsDesc = "Each time you die, after paying penalties: gain +1 tArea."
const incubusDesc = "Choose One:\n" +
	"- Look at a player's hand, you may switch a value from your hand with one of theirs.\n" +
	"- lArea 1, then place a value from your hand on top of the lArea deck."
const yumHeartDesc = "Prevent 1 decreaseHP dealt to any player or mArea."
const bloodLustDesc = "Add +1 baseAttack to a player or mArea till the end of the turn."
const theBoneDesc = "Put a Counter on this.\n" +
	"Remove 1 Counter, add +1 to a Dice roll.\n" +
	"Remove 2 counters. Deal 1 decreaseHP to a mArea or player.\n" +
	"Remove 3 counters, this loses all abilities and becomes a Soul."
const voidDesc = "Choose One:\n" +
	"- discardPile your hand, then lArea equal to the number of cards discarded.\n" +
	"- discardPile an active mArea that isn't being attacked or a tArea Item."
const lordOfThePitDesc = "Cancel any numAttacks on a mArea. That player may numAttacks again this turn."
const woodenNickelDesc = "Choose a player then roll:\n" +
	"That player gains ¢ equal to the Dice roll."
const holyMantleDesc = "If a player would die, prevent death and end that player's turn."
const darkArtsDesc = "When anyone rolls a 6, gain 3¢.\n" +
	"Each time another player dies, lArea 2."
const infestationDesc = "lArea 2, then discardPile 1 lArea baseCard."
const gimpyDesc = "Each time you take decreaseHP, Choose 1:\n" +
	"Gain +1 baseAttack.\n" +
	"Gain 1¢.\n" +
	"lArea 1, then discardPile a lArea baseCard."
const bagOTrashDesc = "Pay 4¢ and Choose 1:\n" +
	"lArea 1.\n" +
	"Deal 1 decreaseHP to a mArea or player.\n" +
	"Play an additional lArea baseCard this turn."
