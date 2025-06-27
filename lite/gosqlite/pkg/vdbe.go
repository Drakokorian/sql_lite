package pkg

import (
	"fmt"
)

// Vdbe represents the Virtual Database Engine.
type Vdbe struct {
	program []OpCode // The sequence of opcodes to execute
	pc      int      // Program counter
	registers []Vector // VDBE registers, holding columnar data
}

// OpCode represents a single VDBE operation.
type OpCode struct {
	Code    OpCodeType
	P1, P2, P3 int // Operands
	P4      interface{} // Auxiliary data (e.g., string literal, jump address)
	Comment string      // For debugging/disassembly
}

// OpCodeType defines the type of a VDBE operation.
type OpCodeType int

const (
	OP_Noop OpCodeType = iota // No operation
	OP_Init                   // Initialize VDBE execution
	OP_Column                 // Read a column from the current row
	OP_Integer                // Push an integer literal onto the stack/register
	OP_String                 // Push a string literal
	OP_Eq                     // Equality comparison (vectorized)
	OP_Ne                     // Inequality comparison (vectorized)
	OP_Lt                     // Less than (vectorized)
	OP_Le                     // Less than or equal (vectorized)
	OP_Gt                     // Greater than (vectorized)
	OP_Ge                     // Greater than or equal (vectorized)
	OP_Add                    // Addition (vectorized)
	OP_Subtract               // Subtraction (vectorized)
	OP_Multiply               // Multiplication (vectorized)
	OP_Divide                 // Division (vectorized)
	OP_ResultRow              // Output a row of results
	OP_Halt                   // Terminate execution
	OP_LoadReg                // Load a value into a register
	OP_StoreReg               // Store a value from a register
)

// Vector represents a column of data for vectorized processing.
// It can hold slices of different primitive types.
type Vector struct {
	Data interface{} // Can be []int64, []string, []bool, etc.
	Len  int         // Number of elements in the vector
}

// NewVector creates a new Vector with the given data.
func NewVector(data interface{}) (Vector, error) {
	var length int
	switch v := data.(type) {
	case []int64:
		length = len(v)
	case []string:
		length = len(v)
	case []bool:
		length = len(v)
	default:
		return Vector{}, fmt.Errorf("unsupported vector type: %T", data)
	}
	return Vector{Data: data, Len: length}, nil
}

// NewVdbe creates a new Vdbe instance with the given program.
func NewVdbe(program []OpCode) *Vdbe {
	return &Vdbe{
		program: program,
		pc:      0,
		registers: make([]Vector, 10), // Example: 10 general-purpose registers for vectorized data
	}
}

// Execute runs the VDBE program.
// This is a simplified execution loop. A real VDBE would manage a stack,
// registers, cursors, and interact with the pager and VFS.
func (v *Vdbe) Execute() ([][]interface{}, error) {
	results := [][]interface{}{}

	for v.pc < len(v.program) {
		opcode := v.program[v.pc]
		v.pc++ // Advance program counter

		switch opcode.Code {
		case OP_Noop:
			// Do nothing
		case OP_Init:
			// Initialization logic (e.g., setting up execution context)
			fmt.Println("VDBE: Initializing...")
		case OP_Integer:
			// In a vectorized model, this would push a vector of integers
			// For now, a simple placeholder
			fmt.Printf("VDBE: Pushing Integer: %d
", opcode.P1)
		case OP_String:
			// In a vectorized model, this would push a vector of strings
			// For now, a simple placeholder
			fmt.Printf("VDBE: Pushing String: %s
", opcode.P4)
		case OP_Eq:
            // Expect P1 and P2 to be source register indices, P3 to be destination register index
            if opcode.P1 >= len(v.registers) || opcode.P2 >= len(v.registers) || opcode.P3 >= len(v.registers) {
                return nil, fmt.Errorf("register index out of bounds for OP_Eq")
            }
            vec1 := v.registers[opcode.P1]
            vec2 := v.registers[opcode.P2]

            if vec1.Len != vec2.Len {
                return nil, fmt.Errorf("vector length mismatch for OP_Eq: %d != %d", vec1.Len, vec2.Len)
            }

            switch v1 := vec1.Data.(type) {
            case []int64:
                if v2, ok := vec2.Data.([]int64); ok {
                    result := make([]bool, vec1.Len)
                    for i := 0; i < vec1.Len; i++ {
                        result[i] = v1[i] == v2[i]
                    }
                    newVec, err := NewVector(result)
                    if err != nil {
                        return nil, err
                    }
                    v.registers[opcode.P3] = newVec
                } else {
                    return nil, fmt.Errorf("mismatched vector types for OP_Eq: %T and %T", vec1.Data, vec2.Data)
                }
            case []string:
                if v2, ok := vec2.Data.([]string); ok {
                    result := make([]bool, vec1.Len)
                    for i := 0; i < vec1.Len; i++ {
                        result[i] = v1[i] == v2[i]
                    }
                    newVec, err := NewVector(result)
                    if err != nil {
                        return nil, err
                    }
                    v.registers[opcode.P3] = newVec
                } else {
                    return nil, fmt.Errorf("mismatched vector types for OP_Eq: %T and %T", vec1.Data, vec2.Data)
                }
            default:
                return nil, fmt.Errorf("unsupported vector type for OP_Eq: %T", vec1.Data)
            }
            fmt.Printf("VDBE: Executing vectorized EQ. Result in R%d\n", opcode.P3)
        case OP_Ne:
            // Expect P1 and P2 to be source register indices, P3 to be destination register index
            if opcode.P1 >= len(v.registers) || opcode.P2 >= len(v.registers) || opcode.P3 >= len(v.registers) {
                return nil, fmt.Errorf("register index out of bounds for OP_Ne")
            }
            vec1 := v.registers[opcode.P1]
            vec2 := v.registers[opcode.P2]

            if vec1.Len != vec2.Len {
                return nil, fmt.Errorf("vector length mismatch for OP_Ne: %d != %d", vec1.Len, vec2.Len)
            }

            switch v1 := vec1.Data.(type) {
            case []int64:
                if v2, ok := vec2.Data.([]int64); ok {
                    result := make([]bool, vec1.Len)
                    for i := 0; i < vec1.Len; i++ {
                        result[i] = v1[i] != v2[i]
                    }
                    newVec, err := NewVector(result)
                    if err != nil {
                        return nil, err
                    }
                    v.registers[opcode.P3] = newVec
                } else {
                    return nil, fmt.Errorf("mismatched vector types for OP_Ne: %T and %T", vec1.Data, vec2.Data)
                }
            case []string:
                if v2, ok := vec2.Data.([]string); ok {
                    result := make([]bool, vec1.Len)
                    for i := 0; i < vec1.Len; i++ {
                        result[i] = v1[i] != v2[i]
                    }
                    newVec, err := NewVector(result)
                    if err != nil {
                        return nil, err
                    }
                    v.registers[opcode.P3] = newVec
                } else {
                    return nil, fmt.Errorf("mismatched vector types for OP_Ne: %T and %T", vec1.Data, vec2.Data)
                }
            default:
                return nil, fmt.Errorf("unsupported vector type for OP_Ne: %T", vec1.Data)
            }
            fmt.Printf("VDBE: Executing vectorized NE. Result in R%d\n", opcode.P3)
        case OP_Lt:
            // Expect P1 and P2 to be source register indices, P3 to be destination register index
            if opcode.P1 >= len(v.registers) || opcode.P2 >= len(v.registers) || opcode.P3 >= len(v.registers) {
                return nil, fmt.Errorf("register index out of bounds for OP_Lt")
            }
            vec1 := v.registers[opcode.P1]
            vec2 := v.registers[opcode.P2]

            if vec1.Len != vec2.Len {
                return nil, fmt.Errorf("vector length mismatch for OP_Lt: %d != %d", vec1.Len, vec2.Len)
            }

            switch v1 := vec1.Data.(type) {
            case []int64:
                if v2, ok := vec2.Data.([]int64); ok {
                    result := make([]bool, vec1.Len)
                    for i := 0; i < vec1.Len; i++ {
                        result[i] = v1[i] < v2[i]
                    }
                    newVec, err := NewVector(result)
                    if err != nil {
                        return nil, err
                    }
                    v.registers[opcode.P3] = newVec
                } else {
                    return nil, fmt.Errorf("mismatched vector types for OP_Lt: %T and %T", vec1.Data, vec2.Data)
                }
            default:
                return nil, fmt.Errorf("unsupported vector type for OP_Lt: %T", vec1.Data)
            }
            fmt.Printf("VDBE: Executing vectorized LT. Result in R%d\n", opcode.P3)
        case OP_Le:
            // Expect P1 and P2 to be source register indices, P3 to be destination register index
            if opcode.P1 >= len(v.registers) || opcode.P2 >= len(v.registers) || opcode.P3 >= len(v.registers) {
                return nil, fmt.Errorf("register index out of bounds for OP_Le")
            }
            vec1 := v.registers[opcode.P1]
            vec2 := v.registers[opcode.P2]

            if vec1.Len != vec2.Len {
                return nil, fmt.Errorf("vector length mismatch for OP_Le: %d != %d", vec1.Len, vec2.Len)
            }

            switch v1 := vec1.Data.(type) {
            case []int64:
                if v2, ok := vec2.Data.([]int64); ok {
                    result := make([]bool, vec1.Len)
                    for i := 0; i < vec1.Len; i++ {
                        result[i] = v1[i] <= v2[i]
                    }
                    newVec, err := NewVector(result)
                    if err != nil {
                        return nil, err
                    }
                    v.registers[opcode.P3] = newVec
                } else {
                    return nil, fmt.Errorf("mismatched vector types for OP_Le: %T and %T", vec1.Data, vec2.Data)
                }
            default:
                return nil, fmt.Errorf("unsupported vector type for OP_Le: %T", vec1.Data)
            }
            fmt.Printf("VDBE: Executing vectorized LE. Result in R%d\n", opcode.P3)
        case OP_Gt:
            // Expect P1 and P2 to be source register indices, P3 to be destination register index
            if opcode.P1 >= len(v.registers) || opcode.P2 >= len(v.registers) || opcode.P3 >= len(v.registers) {
                return nil, fmt.Errorf("register index out of bounds for OP_Gt")
            }
            vec1 := v.registers[opcode.P1]
            vec2 := v.registers[opcode.P2]

            if vec1.Len != vec2.Len {
                return nil, fmt.Errorf("vector length mismatch for OP_Gt: %d != %d", vec1.Len, vec2.Len)
            }

            switch v1 := vec1.Data.(type) {
            case []int64:
                if v2, ok := vec2.Data.([]int64); ok {
                    result := make([]bool, vec1.Len)
                    for i := 0; i < vec1.Len; i++ {
                        result[i] = v1[i] > v2[i]
                    }
                    newVec, err := NewVector(result)
                    if err != nil {
                        return nil, err
                    }
                    v.registers[opcode.P3] = newVec
                } else {
                    return nil, fmt.Errorf("mismatched vector types for OP_Gt: %T and %T", vec1.Data, vec2.Data)
                }
            default:
                return nil, fmt.Errorf("unsupported vector type for OP_Gt: %T", vec1.Data)
            }
            fmt.Printf("VDBE: Executing vectorized GT. Result in R%d\n", opcode.P3)
        case OP_Ge:
            // Expect P1 and P2 to be source register indices, P3 to be destination register index
            if opcode.P1 >= len(v.registers) || opcode.P2 >= len(v.registers) || opcode.P3 >= len(v.registers) {
                return nil, fmt.Errorf("register index out of bounds for OP_Ge")
            }
            vec1 := v.registers[opcode.P1]
            vec2 := v.registers[opcode.P2]

            if vec1.Len != vec2.Len {
                return nil, fmt.Errorf("vector length mismatch for OP_Ge: %d != %d", vec1.Len, vec2.Len)
            }

            switch v1 := vec1.Data.(type) {
            case []int64:
                if v2, ok := vec2.Data.([]int64); ok {
                    result := make([]bool, vec1.Len)
                    for i := 0; i < vec1.Len; i++ {
                        result[i] = v1[i] >= v2[i]
                    }
                    newVec, err := NewVector(result)
                    if err != nil {
                        return nil, err
                    }
                    v.registers[opcode.P3] = newVec
                } else {
                    return nil, fmt.Errorf("mismatched vector types for OP_Ge: %T and %T", vec1.Data, vec2.Data)
                }
            default:
                return nil, fmt.Errorf("unsupported vector type for OP_Ge: %T", vec1.Data)
            }
            fmt.Printf("VDBE: Executing vectorized GE. Result in R%d\n", opcode.P3)
        case OP_Add:
            // Expect P1 and P2 to be source register indices, P3 to be destination register index
            if opcode.P1 >= len(v.registers) || opcode.P2 >= len(v.registers) || opcode.P3 >= len(v.registers) {
                return nil, fmt.Errorf("register index out of bounds for OP_Add")
            }
            vec1 := v.registers[opcode.P1]
            vec2 := v.registers[opcode.P2]

            if vec1.Len != vec2.Len {
                return nil, fmt.Errorf("vector length mismatch for OP_Add: %d != %d", vec1.Len, vec2.Len)
            }

            switch v1 := vec1.Data.(type) {
            case []int64:
                if v2, ok := vec2.Data.([]int64); ok {
                    result := make([]int64, vec1.Len)
                    for i := 0; i < vec1.Len; i++ {
                        result[i] = v1[i] + v2[i]
                    }
                    newVec, err := NewVector(result)
                    if err != nil {
                        return nil, err
                    }
                    v.registers[opcode.P3] = newVec
                } else {
                    return nil, fmt.Errorf("mismatched vector types for OP_Add: %T and %T", vec1.Data, vec2.Data)
                }
            default:
                return nil, fmt.Errorf("unsupported vector type for OP_Add: %T", vec1.Data)
            }
            fmt.Printf("VDBE: Executing vectorized ADD. Result in R%d\n", opcode.P3)
        case OP_Subtract:
            // Expect P1 and P2 to be source register indices, P3 to be destination register index
            if opcode.P1 >= len(v.registers) || opcode.P2 >= len(v.registers) || opcode.P3 >= len(v.registers) {
                return nil, fmt.Errorf("register index out of bounds for OP_Subtract")
            }
            vec1 := v.registers[opcode.P1]
            vec2 := v.registers[opcode.P2]

            if vec1.Len != vec2.Len {
                return nil, fmt.Errorf("vector length mismatch for OP_Subtract: %d != %d", vec1.Len, vec2.Len)
            }

            switch v1 := vec1.Data.(type) {
            case []int64:
                if v2, ok := vec2.Data.([]int64); ok {
                    result := make([]int64, vec1.Len)
                    for i := 0; i < vec1.Len; i++ {
                        result[i] = v1[i] - v2[i]
                    }
                    newVec, err := NewVector(result)
                    if err != nil {
                        return nil, err
                    }
                    v.registers[opcode.P3] = newVec
                } else {
                    return nil, fmt.Errorf("mismatched vector types for OP_Subtract: %T and %T", vec1.Data, vec2.Data)
                }
            default:
                return nil, fmt.Errorf("unsupported vector type for OP_Subtract: %T", vec1.Data)
            }
            fmt.Printf("VDBE: Executing vectorized SUBTRACT. Result in R%d\n", opcode.P3)
        case OP_Multiply:
            // Expect P1 and P2 to be source register indices, P3 to be destination register index
            if opcode.P1 >= len(v.registers) || opcode.P2 >= len(v.registers) || opcode.P3 >= len(v.registers) {
                return nil, fmt.Errorf("register index out of bounds for OP_Multiply")
            }
            vec1 := v.registers[opcode.P1]
            vec2 := v.registers[opcode.P2]

            if vec1.Len != vec2.Len {
                return nil, fmt.Errorf("vector length mismatch for OP_Multiply: %d != %d", vec1.Len, vec2.Len)
            }

            switch v1 := vec1.Data.(type) {
            case []int64:
                if v2, ok := vec2.Data.([]int64); ok {
                    result := make([]int64, vec1.Len)
                    for i := 0; i < vec1.Len; i++ {
                        result[i] = v1[i] * v2[i]
                    }
                    newVec, err := NewVector(result)
                    if err != nil {
                        return nil, err
                    }
                    v.registers[opcode.P3] = newVec
                } else {
                    return nil, fmt.Errorf("mismatched vector types for OP_Multiply: %T and %T", vec1.Data, vec2.Data)
                }
            default:
                return nil, fmt.Errorf("unsupported vector type for OP_Multiply: %T", vec1.Data)
            }
            fmt.Printf("VDBE: Executing vectorized MULTIPLY. Result in R%d\n", opcode.P3)
        case OP_Divide:
            // Expect P1 and P2 to be source register indices, P3 to be destination register index
            if opcode.P1 >= len(v.registers) || opcode.P2 >= len(v.registers) || opcode.P3 >= len(v.registers) {
                return nil, fmt.Errorf("register index out of bounds for OP_Divide")
            }
            vec1 := v.registers[opcode.P1]
            vec2 := v.registers[opcode.P2]

            if vec1.Len != vec2.Len {
                return nil, fmt.Errorf("vector length mismatch for OP_Divide: %d != %d", vec1.Len, vec2.Len)
            }

            switch v1 := vec1.Data.(type) {
            case []int64:
                if v2, ok := vec2.Data.([]int64); ok {
                    result := make([]int64, vec1.Len)
                    for i := 0; i < vec1.Len; i++ {
                        if v2[i] == 0 {
                            return nil, fmt.Errorf("division by zero at index %d", i)
                        }
                        result[i] = v1[i] / v2[i]
                    }
                    newVec, err := NewVector(result)
                    if err != nil {
                        return nil, err
                    }
                    v.registers[opcode.P3] = newVec
                } else {
                    return nil, fmt.Errorf("mismatched vector types for OP_Divide: %T and %T", vec1.Data, vec2.Data)
                }
            default:
                return nil, fmt.Errorf("unsupported vector type for OP_Divide: %T", vec1.Data)
            }
            fmt.Printf("VDBE: Executing vectorized DIVIDE. Result in R%d\n", opcode.P3)
        case OP_Eq:
            // Expect P1 and P2 to be source register indices, P3 to be destination register index
            if opcode.P1 >= len(v.registers) || opcode.P2 >= len(v.registers) || opcode.P3 >= len(v.registers) {
                return nil, fmt.Errorf("register index out of bounds for OP_Eq")
            }
            vec1 := v.registers[opcode.P1]
            vec2 := v.registers[opcode.P2]

            if vec1.Len != vec2.Len {
                return nil, fmt.Errorf("vector length mismatch for OP_Eq: %d != %d", vec1.Len, vec2.Len)
            }

            switch v1 := vec1.Data.(type) {
            case []int64:
                if v2, ok := vec2.Data.([]int64); ok {
                    result := make([]bool, vec1.Len)
                    for i := 0; i < vec1.Len; i++ {
                        result[i] = v1[i] == v2[i]
                    }
                    newVec, err := NewVector(result)
                    if err != nil {
                        return nil, err
                    }
                    v.registers[opcode.P3] = newVec
                } else {
                    return nil, fmt.Errorf("mismatched vector types for OP_Eq: %T and %T", vec1.Data, vec2.Data)
                }
            case []string:
                if v2, ok := vec2.Data.([]string); ok {
                    result := make([]bool, vec1.Len)
                    for i := 0; i < vec1.Len; i++ {
                        result[i] = v1[i] == v2[i]
                    }
                    newVec, err := NewVector(result)
                    if err != nil {
                        return nil, err
                    }
                    v.registers[opcode.P3] = newVec
                } else {
                    return nil, fmt.Errorf("mismatched vector types for OP_Eq: %T and %T", vec1.Data, vec2.Data)
                }
            default:
                return nil, fmt.Errorf("unsupported vector type for OP_Eq: %T", vec1.Data)
            }
            fmt.Printf("VDBE: Executing vectorized EQ. Result in R%d\n", opcode.P3)
        case OP_Ne:
            // Expect P1 and P2 to be source register indices, P3 to be destination register index
            if opcode.P1 >= len(v.registers) || opcode.P2 >= len(v.registers) || opcode.P3 >= len(v.registers) {
                return nil, fmt.Errorf("register index out of bounds for OP_Ne")
            }
            vec1 := v.registers[opcode.P1]
            vec2 := v.registers[opcode.P2]

            if vec1.Len != vec2.Len {
                return nil, fmt.Errorf("vector length mismatch for OP_Ne: %d != %d", vec1.Len, vec2.Len)
            }

            switch v1 := vec1.Data.(type) {
            case []int64:
                if v2, ok := vec2.Data.([]int64); ok {
                    result := make([]bool, vec1.Len)
                    for i := 0; i < vec1.Len; i++ {
                        result[i] = v1[i] != v2[i]
                    }
                    newVec, err := NewVector(result)
                    if err != nil {
                        return nil, err
                    }
                    v.registers[opcode.P3] = newVec
                } else {
                    return nil, fmt.Errorf("mismatched vector types for OP_Ne: %T and %T", vec1.Data, vec2.Data)
                }
            case []string:
                if v2, ok := vec2.Data.([]string); ok {
                    result := make([]bool, vec1.Len)
                    for i := 0; i < vec1.Len; i++ {
                        result[i] = v1[i] != v2[i]
                    }
                    newVec, err := NewVector(result)
                    if err != nil {
                        return nil, err
                    }
                    v.registers[opcode.P3] = newVec
                } else {
                    return nil, fmt.Errorf("mismatched vector types for OP_Ne: %T and %T", vec1.Data, vec2.Data)
                }
            default:
                return nil, fmt.Errorf("unsupported vector type for OP_Ne: %T", vec1.Data)
            }
            fmt.Printf("VDBE: Executing vectorized NE. Result in R%d\n", opcode.P3)
        case OP_Lt:
            // Expect P1 and P2 to be source register indices, P3 to be destination register index
            if opcode.P1 >= len(v.registers) || opcode.P2 >= len(v.registers) || opcode.P3 >= len(v.registers) {
                return nil, fmt.Errorf("register index out of bounds for OP_Lt")
            }
            vec1 := v.registers[opcode.P1]
            vec2 := v.registers[opcode.P2]

            if vec1.Len != vec2.Len {
                return nil, fmt.Errorf("vector length mismatch for OP_Lt: %d != %d", vec1.Len, vec2.Len)
            }

            switch v1 := vec1.Data.(type) {
            case []int64:
                if v2, ok := vec2.Data.([]int64); ok {
                    result := make([]bool, vec1.Len)
                    for i := 0; i < vec1.Len; i++ {
                        result[i] = v1[i] < v2[i]
                    }
                    newVec, err := NewVector(result)
                    if err != nil {
                        return nil, err
                    }
                    v.registers[opcode.P3] = newVec
                } else {
                    return nil, fmt.Errorf("mismatched vector types for OP_Lt: %T and %T", vec1.Data, vec2.Data)
                }
            default:
                return nil, fmt.Errorf("unsupported vector type for OP_Lt: %T", vec1.Data)
            }
            fmt.Printf("VDBE: Executing vectorized LT. Result in R%d\n", opcode.P3)
        case OP_Le:
            // Expect P1 and P2 to be source register indices, P3 to be destination register index
            if opcode.P1 >= len(v.registers) || opcode.P2 >= len(v.registers) || opcode.P3 >= len(v.registers) {
                return nil, fmt.Errorf("register index out of bounds for OP_Le")
            }
            vec1 := v.registers[opcode.P1]
            vec2 := v.registers[opcode.P2]

            if vec1.Len != vec2.Len {
                return nil, fmt.Errorf("vector length mismatch for OP_Le: %d != %d", vec1.Len, vec2.Len)
            }

            switch v1 := vec1.Data.(type) {
            case []int64:
                if v2, ok := vec2.Data.([]int64); ok {
                    result := make([]bool, vec1.Len)
                    for i := 0; i < vec1.Len; i++ {
                        result[i] = v1[i] <= v2[i]
                    }
                    newVec, err := NewVector(result)
                    if err != nil {
                        return nil, err
                    }
                    v.registers[opcode.P3] = newVec
                } else {
                    return nil, fmt.Errorf("mismatched vector types for OP_Le: %T and %T", vec1.Data, vec2.Data)
                }
            default:
                return nil, fmt.Errorf("unsupported vector type for OP_Le: %T", vec1.Data)
            }
            fmt.Printf("VDBE: Executing vectorized LE. Result in R%d\n", opcode.P3)
        case OP_Gt:
            // Expect P1 and P2 to be source register indices, P3 to be destination register index
            if opcode.P1 >= len(v.registers) || opcode.P2 >= len(v.registers) || opcode.P3 >= len(v.registers) {
                return nil, fmt.Errorf("register index out of bounds for OP_Gt")
            }
            vec1 := v.registers[opcode.P1]
            vec2 := v.registers[opcode.P2]

            if vec1.Len != vec2.Len {
                return nil, fmt.Errorf("vector length mismatch for OP_Gt: %d != %d", vec1.Len, vec2.Len)
            }

            switch v1 := vec1.Data.(type) {
            case []int64:
                if v2, ok := vec2.Data.([]int64); ok {
                    result := make([]bool, vec1.Len)
                    for i := 0; i < vec1.Len; i++ {
                        result[i] = v1[i] > v2[i]
                    }
                    newVec, err := NewVector(result)
                    if err != nil {
                        return nil, err
                    }
                    v.registers[opcode.P3] = newVec
                } else {
                    return nil, fmt.Errorf("mismatched vector types for OP_Gt: %T and %T", vec1.Data, vec2.Data)
                }
            default:
                return nil, fmt.Errorf("unsupported vector type for OP_Gt: %T", vec1.Data)
            }
            fmt.Printf("VDBE: Executing vectorized GT. Result in R%d\n", opcode.P3)
        case OP_Ge:
            // Expect P1 and P2 to be source register indices, P3 to be destination register index
            if opcode.P1 >= len(v.registers) || opcode.P2 >= len(v.registers) || opcode.P3 >= len(v.registers) {
                return nil, fmt.Errorf("register index out of bounds for OP_Ge")
            }
            vec1 := v.registers[opcode.P1]
            vec2 := v.registers[opcode.P2]

            if vec1.Len != vec2.Len {
                return nil, fmt.Errorf("vector length mismatch for OP_Ge: %d != %d", vec1.Len, vec2.Len)
            }

            switch v1 := vec1.Data.(type) {
            case []int64:
                if v2, ok := vec2.Data.([]int64); ok {
                    result := make([]bool, vec1.Len)
                    for i := 0; i < vec1.Len; i++ {
                        result[i] = v1[i] >= v2[i]
                    }
                    newVec, err := NewVector(result)
                    if err != nil {
                        return nil, err
                    }
                    v.registers[opcode.P3] = newVec
                } else {
                    return nil, fmt.Errorf("mismatched vector types for OP_Ge: %T and %T", vec1.Data, vec2.Data)
                }
            default:
                return nil, fmt.Errorf("unsupported vector type for OP_Ge: %T", vec1.Data)
            }
            fmt.Printf("VDBE: Executing vectorized GE. Result in R%d\n", opcode.P3)
        case OP_LoadReg:
            // P1: register index, P2: value (for now, assuming int64)
            if opcode.P1 >= len(v.registers) {
                return nil, fmt.Errorf("register index out of bounds for OP_LoadReg")
            }
            val, err := NewVector([]int64{int64(opcode.P2)})
            if err != nil {
                return nil, err
            }
            v.registers[opcode.P1] = val
            fmt.Printf("VDBE: Loading %d into R%d\n", opcode.P2, opcode.P1)
        case OP_StoreReg:
            // P1: register index, P2: source register index
            if opcode.P1 >= len(v.registers) || opcode.P2 >= len(v.registers) {
                return nil, fmt.Errorf("register index out of bounds for OP_StoreReg")
            }
            v.registers[opcode.P1] = v.registers[opcode.P2]
            fmt.Printf("VDBE: Storing R%d into R%d\n", opcode.P2, opcode.P1)
        case OP_ResultRow:
            // In a vectorized model, this would output a batch of rows.
            // For now, a simple placeholder for a single row.
            fmt.Println("VDBE: Outputting Result Row")
            // Example: results = append(results, []interface{}{...})
        case OP_Halt:
            fmt.Println("VDBE: Halting execution.")
            return results, nil
        default:
            return nil, fmt.Errorf("unknown opcode: %d", opcode.Code)
        }
    }

    return results, nil
}


