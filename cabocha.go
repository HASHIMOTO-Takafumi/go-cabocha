package cabocha

// #cgo LDFLAGS: -lcabocha
// #include <stdio.h>
// #include <cabocha.h>
import "C"
import (
	"runtime"
	"unsafe"
)

type Cabocha struct {
	ptr         *C.struct_cabocha_t
	isDestroyed bool
}
type Chunk struct {
	Link           int
	HeadPos        uint
	FuncPos        uint
	TokenSize      uint
	TokenPos       uint
	Score          float32
	Features       []string
	AdditionalInfo string
}
type Token struct {
	Surface           string
	NormalizedSurface string
	Features          []string
	NE                string
	AdditionalInfo    string
	ChunkIndex        uint
}
type Tree struct {
	ptr    *C.struct_cabocha_tree_t
	Chunks []Chunk
	Tokens []Token
}

func NewCabocha(arg string) Cabocha {
	ptr := C.cabocha_new2(C.CString(arg))
	ret := Cabocha{
		ptr:         ptr,
		isDestroyed: false,
	}
	runtime.SetFinalizer(&ret, finalizer)
	return &ret
}

func finalizer(c *Chabocha) {
	if c.isDestroyed == false {
		c.Destroy()
	}
	return
}

func (c *Cabocha) Parse(str string) Tree {
	ptr := C.cabocha_sparse_totree(c.ptr, C.CString(str))

	nChunks := uint(C.cabocha_tree_chunk_size(ptr))
	chunks := make([]Chunk, nChunks)
	chunkByPtr := map[*C.struct_cabocha_chunk_t]uint{}
	for i := uint(0); i < nChunks; i++ {
		chunk := C.cabocha_tree_chunk(ptr, C.size_t(i))
		chunks[i] = Chunk{
			Link:           int(chunk.link),
			HeadPos:        uint(chunk.head_pos),
			FuncPos:        uint(chunk.func_pos),
			TokenSize:      uint(chunk.token_size),
			TokenPos:       uint(chunk.token_pos),
			Score:          float32(chunk.score),
			Features:       readStrings(chunk.feature_list, chunk.feature_list_size),
			AdditionalInfo: C.GoString(chunk.additional_info),
		}
		chunkByPtr[chunk] = i
	}

	nToken := uint(C.cabocha_tree_token_size(ptr))
	tokens := make([]Token, nToken)
	for i := uint(0); i < nToken; i++ {
		token := C.cabocha_tree_token(ptr, C.size_t(i))
		tokens[i] = Token{
			Surface:           C.GoString(token.surface),
			NormalizedSurface: C.GoString(token.normalized_surface),
			Features:          readStrings(token.feature_list, token.feature_list_size),
			NE:                C.GoString(token.ne),
			AdditionalInfo:    C.GoString(token.additional_info),
			ChunkIndex:        chunkByPtr[token.chunk],
		}
	}

	return Tree{
		ptr:    ptr,
		Chunks: chunks,
		Tokens: tokens,
	}
}

func (c *Cabocha) Destroy() {
	c.isDestroyed = true
	C.cabocha_destroy(c.ptr)
}

func readStrings(p **C.char, n C.ushort) []string {
	list := make([]string, n)

	ptr := uintptr(unsafe.Pointer(p))
	d := unsafe.Sizeof(C.size_t(0))
	for i := C.ushort(0); i < n; i++ {
		list[i] = C.GoString(*(**C.char)(unsafe.Pointer(ptr + uintptr(i)*d)))
	}

	return list
}
