package calc

// Stack is a LIFO data structure
type Stack struct {
	Values []Token
}

// Pop removes the token at the top of the stack and returns its value
func (self *Stack) Pop() Token {
	if len(self.Values) == 0 {
		return Token{}
	}
	token := self.Values[len(self.Values)-1]
	self.Values = self.Values[:len(self.Values)-1]
	return token
}

// Push adds tokens to the top of the stack
func (self *Stack) Push(i ...Token) {
	self.Values = append(self.Values, i...)
}

// Peek returns the token at the top of the stack
func (self *Stack) Peek() Token {
	if len(self.Values) == 0 {
		return Token{}
	}
	return self.Values[len(self.Values)-1]
}

// EmptyInto dumps all tokens from one stack to another
func (self *Stack) EmptyInto(s *Stack) {
	if !self.IsEmpty() {
		for i := self.Length() - 1; i >= 0; i-- {
			s.Push(self.Pop())
		}
	}
}

// IsEmpty checks if there are any tokens in the stack
func (self *Stack) IsEmpty() bool {
	return len(self.Values) == 0
}

// Length returns the amount of tokens in the stack
func (self *Stack) Length() int {
	return len(self.Values)
}
