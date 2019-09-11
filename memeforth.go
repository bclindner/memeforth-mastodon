package main

import (
	"fmt"
	"unicode"
)

// Base type for a Forth stack, containing functions which allow it to carry
// out its base functionalities.
type ForthStack []string

// ForthStack helper functions
func (stack *ForthStack) pop() string {
	el := (*stack)[len(*stack) - 1]
	*stack = (*stack)[:len(*stack) - 1]
	return el
}
func (stack *ForthStack) push(el string) {
	*stack = append(*stack, el)
}

func (stack *ForthStack) Concat() (err error) {
	// Verify stack is big enough for us to use
	if len(*stack) < 2 {
		return fmt.Errorf("CONCAT requires 2 tokens")
	}
	// Get the two strings, add them, and append it to the sliced stack
	stringOne := stack.pop()
	stringTwo := stack.pop()
	cattedString := stringTwo + stringOne
	stack.push(cattedString)
	return nil
}

func (stack *ForthStack) TokenEnhance(format string) (err error) {
	// Verify stack has one token
	if len(*stack) < 1 {
		return fmt.Errorf("DNS requires 1 token")
	}
	subject := stack.pop()
	dnsOverSubject := fmt.Sprintf(format, subject)
	stack.push(dnsOverSubject)
	return nil
}

func (stack *ForthStack) Emojify(emojiname string) (err error) {
	// Verify stack has one token
	if len(*stack) < 1 {
		return fmt.Errorf("DNS requires 1 token")
	}
	// Lord forgive me for what I'm bout to do
	var newToken string
	token := stack.pop()
	for _, char := range token {
		if unicode.IsLetter(char) || unicode.IsNumber(char) {
			newToken += ":" + emojiname + "_" + string(unicode.ToLower(char)) + ": "
		} else {
			// Just add unparseable tokens in as themselves, I guess
			newToken += string(char)
		}
	}
	stack.push(newToken)
	return nil
}

// Rudimentary Forth parser, originally written by @mdszy@mastodon.technology
// and ported to Go.
func ProcessMemeForth(code string) (output string, err error) {
	// Tokenize

	// Whether or not the current character is part of a literal string
	var litstring bool
	// The current token being read
	var token string
	// The slice of tokens that have already been read
	var tokens []string
	for _, char := range code {
		switch char {
		// if this is a single quote, we are starting or ending a string literal
		case '\'':
			// start the string literal if not started
			if !litstring {
				litstring = true
				break
			}
			// if already started, end the string literal and push it as a token
			litstring = false
			if len(token) > 0 {
				tokens = append(tokens, token)
				token = ""
			}
			break
		// if this is a space, and we are not in a literal, this marks the end of a token
		// if we are in a literal, then simply add the character to the literal
		case ' ':
			if litstring {
				token = token + string(char)
				break
			}
			if len(token) > 0 {
				tokens = append(tokens, token)
				token = ""
			}
			break
		// if this is any other character, add it to the token
		default:
			token = token + string(char)
			break
		}
	}
	// if there is still a token, then it is the final token - add it
	if len(token) > 0 {
		tokens = append(tokens, token)
	}
	
	// Initialize stack, and run the code
	var stack ForthStack
	for _, token := range tokens {
		switch token {
		case "CONCAT":
			err = stack.Concat()
			break
		case "DNS":
			err = stack.TokenEnhance("DNS over %v")
			break
		case "SUREHOPE":
			err = stack.TokenEnhance("%v? I sure hope it does!")
			break
		case "HARDWORK":
			err = stack.TokenEnhance("Another hard day's work at the '%v' factory.")
			break
		case "SM64":
			err = stack.Emojify("sm64")
			break
		case "HACKER":
			err = stack.Emojify("hacker")
			break
		default:
			stack = append(stack, token)
		}
		if err != nil {
			return "", err
		}
	}
	return stack[len(stack) - 1], nil
}
