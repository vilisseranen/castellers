#!/usr/bin/env python

import os
import json

BASE_PATH = "../../src/assets/translations/"

def main():
    translations = load_translations()
    mssing_translations = check_missing_keys(translations)

def load_translations():
    translations = []
    for filename in os.listdir(BASE_PATH):
        with open(BASE_PATH + filename) as f:
            translations.append(json.loads(f.read()))
    return translations

def check_missing_keys(translations):
    missing_translations = {}
    all_translations = {}
    count_translated = { 'fr': 0, 'en': 0, 'cat': 0 }
    total_translations = 0
    # for each file
    for translation in translations:
        # for each language in a file
        for language in translation:
            # First time we see this language
            if language not in missing_translations:
                missing_translations[language] = {}
            # for each category (there should be only one)
            for category in translation[language]:
                # First time we see this category
                if category not in all_translations:
                    all_translations[category] = set()
                # Get all keys for all translations
                all_translations[category].update(translation[language][category].keys())
    for translation in translations:
        for language in translation:
            for category in translation[language]:
                for item in (all_translations[category] - set(translation[language][category].keys())):
                    count_translated[language] += 1.0
                    print "missing: {}.{}.{}".format(language, category, item)
    for category in all_translations:
        total_translations += len(all_translations[category])
    print
    for language in count_translated:
        print "{} {:.0f}% translated".format(language, (total_translations-count_translated[language])/total_translations*100)

if __name__ == "__main__":
    main()
