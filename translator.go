package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"strings"
	"path"
)

type Translation struct {
	Template   string
}

type Language map[string]Translation

type Translator struct {
	directory    string
	languagesRaw map[string][]byte
	languages    map[string]Language
	mutexRaw        sync.Mutex
	mutexLang        sync.Mutex
}

func NewTranslator(directory string) (*Translator, error) {

	// check if directory exists
	info, err := os.Stat(directory)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is no directory", directory)
	}
	t := &Translator{
		directory:    directory,
		languagesRaw: make(map[string][]byte),
		languages:    make(map[string]Language),
	}
	// load translations
	if err := t.Load(); err != nil {
		return nil, err
	}
	return t, nil
}

// Load all translations from the translation directory
func (t *Translator) Load() error {
	// iterate over directory
	return filepath.Walk(t.directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// ignore dirs
		if info.IsDir() {
			return nil
		}
		// match json files
		if matched, _ := filepath.Match("*.json", info.Name()); matched {
			// open file
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			// read all bytes
			b, err := ioutil.ReadAll(file)
			if err != nil {
				return err
			}
			name := strings.Replace(info.Name(), ".json", "", -1)
			fmt.Println(name)
			// add language
			err = t.parseLanguage(name, b)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// Get returns the value of a key
func (t *Translator) Get(key, lang string) string {
	// find language
	if trans, ok := t.languages[lang]; ok {
		if value, ok := trans[key]; ok {
			return value.Template
		}
	}
	return key
}

// Set the value for a key
func (t *Translator) Set(key, value, lang string) error {
	// find language
	if trans, ok := t.languages[lang]; ok {
		// update key
		t.mutexLang.Lock()
		entry := trans[key]
		entry.Template = value
		trans[key] = entry
		t.mutexLang.Unlock()
		return nil
	}
	return fmt.Errorf("Language %s is not loaded", lang)
}

// Sync
func (t *Translator) Sync(base string, orphanRemoval bool) error {

	baseLanguage, ok := t.languages[base]
	if !ok {
		return fmt.Errorf("Language %s is not loaded", base)
	}
	// iterate over all languages and remove key
	if orphanRemoval {
		for langCode, language := range t.languages {
			if langCode != base {
				for key := range language {
					if _, ok := baseLanguage[key]; !ok {
						t.mutexLang.Lock()
						delete(language, key)
						t.mutexLang.Unlock()
					}
				}
			}
		}
	}
	// sync
	for key, translation := range baseLanguage {
		for langCode, language := range t.languages {
			if langCode != base {
				// check if key exists in language
				if _, ok := language[key]; !ok {
					t.mutexLang.Lock()
					language[key] = translation
					t.mutexLang.Unlock()
				}
			}
		}
	}

	return nil
}

// Save all changes to file
func (t *Translator) Save(indent bool) error {
	for lang := range t.languages {
		if err := t.syncRaw(lang, indent); err != nil {
			return err
		}
		// save to file
		f, err := os.Create(path.Join(t.directory, lang + ".json"))
		if err != nil {
			return err
		}
		_, err = f.Write(t.languagesRaw[lang])
		if err != nil {
			return err
		}
		err = f.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Translator) parseLanguage(name string, b []byte) error {
	var lang map[string]interface{}
	err := json.Unmarshal(b, &lang)
	if err != nil {
		return err
	}
	// add language to map
	t.mutexRaw.Lock()
	t.languagesRaw[name] = b
	t.mutexRaw.Unlock()
	// add optimized translations
	t.mutexLang.Lock()
	t.languages[name] = t.extractKeys("", lang)
	t.mutexLang.Unlock()

	return nil
}

func (t *Translator) extractKeys(prefix string, m map[string]interface{}) map[string]Translation {
	if prefix != "" {
		prefix += "."
	}
	keys := make(map[string]Translation)
	for k, v := range m {
		key := prefix + k
		switch v.(type) {
		case string:
			keys[key] = Translation{Template: v.(string)}
		case map[string]interface{}:
			sub := t.extractKeys(key, v.(map[string]interface{}))
			// merge
			for sk, vk := range sub {
				keys[sk] = vk
			}
		}
	}
	return keys
}

func (t *Translator) syncRaw(lang string, indent bool) (err error) {
	if language, ok := t.languages[lang]; ok {
		target := make(map[string]interface{})
		for key, translation := range language {
			insert(key, translation.Template, target)
		}
		var b []byte
		if indent {
			b, err = json.MarshalIndent(target, "", "  ")
		} else {
			b, err = json.Marshal(target)
		}
		if err != nil {
			return err
		}
		// update raw language
		t.mutexRaw.Lock()
		t.languagesRaw[lang] = b
		t.mutexRaw.Unlock()
		return nil
	}
	return fmt.Errorf("Language %s is not loaded", lang)
}

func insert(key, value string, target map[string]interface{}) {
	if !strings.Contains(key, ".") {
		target[key] = value
		return
	}
	keyParts := strings.SplitN(key, ".", 2)
	// check if child exists in target
	var child map[string]interface{}
	if c, ok := target[keyParts[0]]; ok {
		// add child to target
		child = c.(map[string]interface{})
	} else {
		child = make(map[string]interface{})
	}
	// insert child
	insert(keyParts[1], value, child)
	target[keyParts[0]] = child
}