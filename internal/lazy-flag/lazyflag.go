package lazyflag

// func BindFlag(dst *pflag.FlagSet, v any) error {
// 	return bind_flag(dst, reflect.ValueOf(v))
// }

// func bind_flag(dst *pflag.FlagSet, v reflect.Value) error {
// 	vtp := v.Type()
// 	if vtp.Kind() == reflect.Pointer {
// 		vtp = vtp.Elem()
// 		v = v.Elem()
// 	}
// 	fieldcount := vtp.NumField()
// 	for i := 0; i < fieldcount; i++ {
// 		field := vtp.Field(i)
// 		fieldv := v.Field(i)
// 		kind := field.Type.Kind()
// 		switch kind{
// 		case reflect.Int:
// 			dst.IntVar()
// 		}
// 	}
// }
