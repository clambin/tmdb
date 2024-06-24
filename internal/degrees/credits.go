package degrees

type actorCredits map[int]string

func (a actorCredits) add(id int, title string) {
	a[id] = title
}

func commonActorCredits(a, b actorCredits) actorCredits {
	common := make(actorCredits, len(a))
	for idA, titleA := range a {
		if _, ok := b[idA]; ok {
			common[idA] = titleA
		}
	}
	return common
}
