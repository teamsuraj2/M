package database

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"main/config"
)

const (
	nsfwWordDocID = "badwords"
)

func AddNSFWWord(word string) error {
	word = strings.ToLower(word)

	existing, err := GetNSFWWords()
	if err != nil {
		return err
	}

	for _, w := range existing {
		if w == word {
			return nil
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err = nsfwWordsDB.UpdateOne(
		ctx,
		bson.M{"_id": nsfwWordDocID},
		bson.M{"$addToSet": bson.M{"words": word}},
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		return err
	}

	existing = append(existing, word)
	config.Cache.Store("nsfw", existing)
	return nil
}

func RemoveNSFWWord(word string) error {
	word = strings.ToLower(word)

	existing, err := GetNSFWWords()
	if err != nil {
		return err
	}

	updated := make([]string, 0, len(existing))
	found := false
	for _, w := range existing {
		if w != word {
			updated = append(updated, w)
		} else {
			found = true
		}
	}

	if !found {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err = nsfwWordsDB.UpdateOne(
		ctx,
		bson.M{"_id": nsfwWordDocID},
		bson.M{"$pull": bson.M{"words": word}},
	)
	if err != nil {
		return err
	}

	config.Cache.Store("nsfw", updated)
	return nil
}


func GetNSFWWords() ([]string, error) {
	if val, ok := config.Cache.Load("nsfw"); ok {
		if words, ok := val.([]string); ok {
			return words, nil
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	BanWords := []string{
		"s*x", "n*de", "f*ck", "f?ck", "b*tch", "d*ck", "p*ssy", "c*ck", "a*shole", "hot*girl", "hot**girl",
		"p*rn", "sl*t", "b*obs",
		// Hindi (Devanagari script)
		"रंडी", "चोद", "मादरचोद", "गांड", "लंड", "भोसड़ी", "हिजड़ा", "पागल", "नंगा",

		"ch*tiya", "m*derchod", "b*henchod", "g*ndu", "r*ndi", "b*osdi", "hijda",
		"l*nd", "ch*d", "jh*tu", "h*rami", "k*mina", "sa*la", "g*nd", "p*gal",
		"bh*dwa", "ch*t", "h*ramkhor", "ch*lu", "g*dha", "b*dtameez", "k*njoos",
		"ch*pri", "s*st", "ull*", "k*ttiya",

		// Phrases
		"tera baap", "teri maa", "teri behan", "maa ka bhosda", "gaand maar dunga",

		"t*ra baap", "t*ri maa", "t*ri behan", "maa ka b*osda", "g*nd maar d*nga", "d*epthroat", "h*ntai", "b*dsm", "l*sbian", "f*ta", "cam*girl", "call*girl", "sex*chat", "child*porn", "p*do", "teen*sex", "casting*couch", "strip*club", "only*fans", "bikini*photos", "lingam*massage", "tantra*sex", "lick*pussy", "tight*pussy", "wet*pussy", "h*ndjob", "cleavage**show", "massage**sex", "body**massage", "bathroom*sex", "desi*call*girl",
		// Wildcard-based patterns for obfuscation handling
		"f?ck", "f*ck", "b*tch", "p*ssy", "d*ck", "l*nd", "g*nd", "hot**video", "ch*da*i", "r*ndi",

		"aad", "a*ad",
		"aand", "aa*d",
		"bahenchod", "bahen*chod", "ba*enchod",
		"behenchod", "behen*chod", "beh*nchod",
		"bhenchod", "bhen*chod",
		"bhenchodd", "bhen*chodd",
		"b.c.", "b?c?", "b*c*",
		"bc", "b?c", "b*c",
		"bakchod", "bak*chod", "ba*kchod",
		"bakchodd", "bak*chodd",
		"bakchodi", "bak*chodi",
		"bevda", "bev*da", "b?vda",
		"bewda", "bew*da",
		"bevdey", "bev*dey",
		"bewday", "bew*day",
		"bevakoof", "beva*koof",
		"bevkoof", "bev*koof",
		"bevkuf", "bev*kuf",
		"bewakoof", "bewa*koof",
		"bewkoof", "bew*koof",
		"bewkuf", "bew*kuf",
		"bhadua", "bha*dua", "bh?dua",
		"bhadvaa", "bhad*vaa",
		"bhadwa", "bha*dwa",
		"bhosada", "bhos*ada",
		"bhosda", "bhos*da", "b?osda",
		"bhosdaa", "bhos*daa",
		"bhosdike", "bhos*dike", "bhos?ike",
		"bhonsdike", "bhons*dike",
		"bsdk", "b*dk", "bs?k", "b?sdk",
		"b.s.d.k", "b*s*d*k",
		"bhosdiki", "bhos*diki",
		"bhosdiwala", "bhos*wala", "bhos?wala",
		"bhosdiwale", "bhos*diwale",
		"bhosadchod", "bhos*chod",
		"bhosadchodal", "bhos*chodal",
		"babbe", "bab*e",
		"babbey", "bab*ey",
		"bube", "bu*e",
		"bubey", "bu*bey",
		"bur", "b*r",
		"burr", "bu*rr",
		"buurr", "bu*urr",
		"buur", "bu*ur",
		"charsi", "char*si",
		"chooche", "choo*che",
		"choochi", "choo*chi",
		"chuchi", "chu*chi",
		"chhod", "chh*d",
		"chod", "ch*d", "ch?d", "chut*", "chu*", "chu?t", "chu?t*",
		"chodd", "ch*odd",
		"chudne", "chud*ne",
		"chudney", "chud*ney",
		"chudwa", "chud*wa",
		"chudwaa", "chud*waa",
		"chudwane", "chud*wane",
		"chudwaane", "chud*waane",
		"choot", "ch*t", "cho*t", "ch?ot", "chut*", "chu*t", "chu*t*",
		"chut", "ch*t", "chu*", "chut*", "chut**", "chut?", "chu?t",
		"chute", "chu*te",
		"chutia", "ch*tia",
		"chutiya", "chu*iya", "ch?tiya",
		"chutiye", "chu*iye",
		"chuttad", "chu*ttad",
		"chutad", "chu*tad",
		"dalaal", "dal*al",
		"dalal", "dal*l",
		"dalle", "dal*le",
		"dalley", "dal*ley",
		"fattu", "fat*tu",
		"gadhalund", "gadh*lund",
		"gaand", "ga*nd", "g*nd",
		"gand", "g*nd",
		"gandu", "gan*du",
		"gandfat", "gand*fat",
		"gandfut", "gand*fut",
		"gandiya", "gan*diya",
		"gandiye", "gan*diye",
		"goo", "g*o",
		"gu", "g*u",
		"gote", "go*te",
		"gotey", "go*tey",
		"gotte", "go*tte",
		"hag", "h*g",
		"haggu", "hag*gu",
		"hagne", "hag*ne",
		"hagney", "hag*ney",
		"harami", "ha*rami",
		"haramjada", "har*amjada",
		"haraamjaada", "hara*amjaada",
		"haramzyada", "haram*zyada",
		"haraamzyaada", "haraam*zyaada",
		"haraamjaade", "haraam*jaade",
		"haraamzaade", "haraam*zaade",
		"haraamkhor", "har?mkhor", "haraam*khor",
		"haramkhor", "har*mkhor",
		"jhat", "j*hat",
		"jhaat", "jhaa*t",
		"jhaatu", "jhaa*tu",
		"jhatu", "jha*tu",
		"kutta", "kut*ta",
		"kutte", "kut*te",
		"kuttey", "kut*tey",
		"kutia", "kut*ia",
		"kutiya", "kut*iya",
		"kuttiya", "kut*tiya",
		"kutti", "kut*ti",
		"landi", "lan*di",
		"landy", "lan*dy",
		"laude", "lau*de",
		"laudey", "lau*dey",
		"lauda", "lau*da",
		"lora", "lo*ra",
		"laura", "lau*ra",
		"loda", "lo*da",
		"lode", "lo*de",
		"lulli", "lu*lli",
		"ling", "l*ng",
		"loda", "lo*da",
		"lode", "lo*de",
		"lund", "l*nd", "lu*d", "l?nd",
		"launda", "lau*nda",
		"lounde", "lou*nde",
		"laundey", "lau*ndey",
		"laundi", "lau*ndi",
		"loundi", "lou*ndi",
		"laundiya", "lau*ndiya",
		"loundiya", "lou*ndiya",
		"maar", "ma*ar",
		"maro", "ma*ro",
		"marunga", "ma*runga",
		"madarchod", "m?darchod", "madar*chod",
		"madarchodd", "madar*chodd",
		"madarchood", "madar*chood",
		"madarchoot", "madar*choot",
		"madarchut", "madar*chut",
		"m.c.", "m*c*", "m?c?",
		"mc", "m*c", "m?c",
		"mamme", "mam*me",
		"mammey", "mam*mey",
		"moot", "mo*t",
		"mut", "m*ut",
		"mootne", "moot*ne",
		"mutne", "mut*ne",
		"mooth", "mo*oth",
		"muth", "mu*th",
		"nunni", "nu*nni",
		"nunnu", "nu*nnu",
		"paaji", "paa*ji",
		"paji", "pa*ji",
		"pesaab", "pe*saab",
		"pesab", "pe*sab",
		"peshaab", "pe*shaab",
		"peshab", "pe*shab",
		"pilla", "pi*lla",
		"pillay", "pi*llay",
		"pille", "pi*lle",
		"pilley", "pi*lley",
		"pisaab", "pi*saab",
		"pisab", "pi*sab",
		"pkmkb", "pkm*kb",
		"porkistan", "por*kistan",
		"raand", "raa*nd",
		"rand", "ra*nd",
		"randi", "ran*di", "r*ndi",
		"randy", "ran*dy",
		"suar", "su*ar",
		"tatte", "ta*tte",
		"tatti", "ta*tti",
		"tatty", "ta*tty",
		"ullu", "ul*lu",

		"आं?ड*", "आ*ंड", "आ*ँड",
		"बह*नचोद", "बे*हेनचोद", "भे*नचोद",
		"ब*कचोद", "ब*कचोदी",
		"बेव*ड़ा", "बेव*ड़े", "बेव*कूफ",
		"भड़*आ", "भड़*वा",
		"भोस*ड़ा", "भोस*ड़ीके", "भोस*ड़ीकी", "भोस*ड़ीवाला", "भोस*ड़ीवाले",
		"भोसर*चोदल", "भोसद*चोद", "भोस*ड़ाचोदल", "भोस*ड़ाचोद",
		"बब*बे", "बू*बे", "बु*र",
		"च*रसी", "चू*चे", "चू*ची", "चु*ची",
		"चोद*", "चुद*ने", "चुद*वा", "चुद*वाने",
		"चू*त", "चू*तिया", "चु*टिया", "चू*तिये", "चुत्त*ड़", "चूत्त*ड़",
		"दला*ल", "दल*ले",
		"फट*टू",
		"गध*ा", "गध*े", "गधा*लंड",
		"गां*ड", "गां*डू", "गं*डफट", "गं*डिया", "गं*डिये",
		"गू*",
		"गो*टे",
		"हग*", "हग*गू", "हग*ने",
		"हराम*ी", "हराम*जादा", "हराम*ज़ादा", "हराम*जादे", "हराम*ज़ादे", "हराम*खोर",
		"झा*ट", "झा*टू",
		"कुत*ता", "कुत*ते", "कुत*िया", "कुत*ती",
		"लें*डी", "लो*ड़े", "लौ*ड़े", "लौ*ड़ा", "लो*ड़ा", "लौ*डा",
		"लिं*ग", "लो*डा", "लो*डे", "लं*ड",
		"लौ*ंडा", "लौ*ंडे", "लौ*ंडी", "लौ*ंडिया",
		"लु*ल्ली",
		"मा*र", "मा*रो", "मा*रूंगा",
		"मादर*चोद", "मादर*चूत", "मादर*चुत",
		"मम्म*े",
		"मू*त", "मु*त", "मू*तने", "मु*तने", "मू*ठ", "मु*ठ",
		"नुन*नी", "नु*नु",
		"पा*जी",
		"पे*साब", "पे*शाब",
		"पिल*ला", "पिल*ले",
		"पिस*ाब",
		"पो*रकिस्तान",
		"रा*ंड", "रं*डी",
		"सु*अर", "सू*अर",
		"ट*ट्टे", "ट*ट्टी",
	}

	
	var result struct {
		Words []string `bson:"words"`
	}

	err := nsfwWordsDB.FindOne(ctx, bson.M{"_id": nsfwWordDocID}).Decode(&result)
	if err != nil && err != mongo.ErrNoDocuments {
		return BanWords, err
	}

	unique := make(map[string]struct{})
	for _, word := range BanWords {
		unique[word] = struct{}{}
	}
	for _, word := range result.Words {
		unique[word] = struct{}{}
	}

	combined := make([]string, 0, len(unique))
	for word := range unique {
		combined = append(combined, word)
	}

	config.Cache.Store("nsfw", combined)
	return combined, nil
}


func SetNSFWFlag(chatID int64, enable bool) error {
	cacheKey := fmt.Sprintf("%d_nsfw", chatID)

	if cached, ok := config.Cache.Load(cacheKey); ok {
		if cached.(bool) == enable {
			return nil
		}
	}

	old, err := IsNSFWEnabled(chatID)
	if err != nil {
		return err
	}
	if old == enable {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err = nsfwFlagsDB.UpdateOne(
		ctx,
		bson.M{"_id": chatID},
		bson.M{"$set": bson.M{"enabled": enable}},
		options.UpdateOne().SetUpsert(true),
	)
	if err == nil {
		config.Cache.Store(cacheKey, enable)
	}
	return err
}

func IsNSFWEnabled(chatID int64) (bool, error) {
	key := fmt.Sprintf("%d_nsfw", chatID)

	if val, ok := config.Cache.Load(key); ok {
		if enabled, ok := val.(bool); ok {
			return enabled, nil
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var result struct {
		Enabled bool `bson:"enabled"`
	}

	err := nsfwFlagsDB.FindOne(ctx, bson.M{"_id": chatID}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			config.Cache.Store(key, false)
			return false, nil
		}
		return false, err
	}

	config.Cache.Store(key, result.Enabled)
	return result.Enabled, nil
}
