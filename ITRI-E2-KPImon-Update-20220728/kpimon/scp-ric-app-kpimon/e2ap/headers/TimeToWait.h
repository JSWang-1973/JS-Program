/*
 * Generated by asn1c-0.9.29 (http://lionet.info/asn1c)
 * From ASN.1 module "E2AP-IEs"
 * 	found in "E2AP-IEs-v01.00.asn"
 * 	`asn1c -pdu=auto -fno-include-deps -fcompound-names -findirect-choice -gen-PER -gen-OER -no-gen-example`
 */

#ifndef	_TimeToWait_H_
#define	_TimeToWait_H_


#include <asn_application.h>

/* Including external dependencies */
#include <NativeEnumerated.h>

#ifdef __cplusplus
extern "C" {
#endif

/* Dependencies */
typedef enum TimeToWait {
	TimeToWait_v1s	= 0,
	TimeToWait_v2s	= 1,
	TimeToWait_v5s	= 2,
	TimeToWait_v10s	= 3,
	TimeToWait_v20s	= 4,
	TimeToWait_v60s	= 5
	/*
	 * Enumeration is extensible
	 */
} e_TimeToWait;

/* TimeToWait */
typedef long	 TimeToWait_t;

/* Implementation */
extern asn_per_constraints_t asn_PER_type_TimeToWait_constr_1;
extern asn_TYPE_descriptor_t asn_DEF_TimeToWait;
extern const asn_INTEGER_specifics_t asn_SPC_TimeToWait_specs_1;
asn_struct_free_f TimeToWait_free;
asn_struct_print_f TimeToWait_print;
asn_constr_check_f TimeToWait_constraint;
ber_type_decoder_f TimeToWait_decode_ber;
der_type_encoder_f TimeToWait_encode_der;
xer_type_decoder_f TimeToWait_decode_xer;
xer_type_encoder_f TimeToWait_encode_xer;
oer_type_decoder_f TimeToWait_decode_oer;
oer_type_encoder_f TimeToWait_encode_oer;
per_type_decoder_f TimeToWait_decode_uper;
per_type_encoder_f TimeToWait_encode_uper;
per_type_decoder_f TimeToWait_decode_aper;
per_type_encoder_f TimeToWait_encode_aper;

#ifdef __cplusplus
}
#endif

#endif	/* _TimeToWait_H_ */
#include <asn_internal.h>
