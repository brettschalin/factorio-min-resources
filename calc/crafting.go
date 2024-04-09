package calc

func countWithBonus(wanted int, bonus float64, targetIngs bool) (ing, prod int) {

	if bonus < 0 {
		panic("productivity bonus must be >= 0")
	}

	if bonus == 0 {
		return wanted, wanted
	}

	var b float64

	for (targetIngs && ing < wanted) || (!targetIngs && prod < wanted) {
		ing += 1
		prod += 1
		b += bonus
		if b >= 1 {
			b -= 1
			prod += 1
		}
	}

	return
}
