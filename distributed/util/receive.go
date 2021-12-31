package util


func Receive(h,w int, cell []Cell)[][]byte{
	world2 := make([][]byte, 5120)
	for i := range world2 {
		world2[i] = make([]byte, 5120)
	}
	for _,i := range(cell){
		world2[i.X][i.Y] = 255
	}
	return world2
}