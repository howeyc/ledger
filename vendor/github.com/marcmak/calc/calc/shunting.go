package calc

func ShuntingYard(s Stack) Stack {
	postfix := Stack{}
	operators := Stack{}
	for _, v := range s.Values {
		switch v.Type {
		case OPERATOR:
			for !operators.IsEmpty() {
				val := v.Value
				top := operators.Peek().Value
				if (oprData[val].prec <= oprData[top].prec && oprData[val].rAsoc == false) ||
					(oprData[val].prec < oprData[top].prec && oprData[val].rAsoc == true) {
					postfix.Push(operators.Pop())
					continue
				}
				break
			}
			operators.Push(v)
		case LPAREN:
			operators.Push(v)
		case RPAREN:
			for i := operators.Length() - 1; i >= 0; i-- {
				if operators.Values[i].Type != LPAREN {
					postfix.Push(operators.Pop())
					continue
				} else {
					operators.Pop()
					break
				}
			}
		default:
			postfix.Push(v)
		}
	}
	operators.EmptyInto(&postfix)
	return postfix
}
