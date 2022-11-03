#include <errno.h>
#include "wrapper.h"

ssize_t e2sm_encode_ric_event_trigger_definition(void *buffer, size_t buf_size, size_t event_trigger_count, long *RT_periods) {
	E2SM_KPM_EventTriggerDefinition_t *eventTriggerDef = (E2SM_KPM_EventTriggerDefinition_t *)calloc(1, sizeof(E2SM_KPM_EventTriggerDefinition_t));
	if(!eventTriggerDef) {
		fprintf(stderr, "alloc EventTriggerDefinition failed\n");
		return -1;
	}

	E2SM_KPM_EventTriggerDefinition_Format1_t *innerDef = (E2SM_KPM_EventTriggerDefinition_Format1_t *)calloc(1, sizeof(E2SM_KPM_EventTriggerDefinition_Format1_t));
	if(!innerDef) {
		fprintf(stderr, "alloc EventTriggerDefinition Format1 failed\n");
		ASN_STRUCT_FREE(asn_DEF_E2SM_KPM_EventTriggerDefinition, eventTriggerDef);
		return -1;
	}

	eventTriggerDef->present = E2SM_KPM_EventTriggerDefinition_PR_eventDefinition_Format1;
	eventTriggerDef->choice.eventDefinition_Format1 = innerDef;

	struct E2SM_KPM_EventTriggerDefinition_Format1__policyTest_List *policyTestList = (struct E2SM_KPM_EventTriggerDefinition_Format1__policyTest_List *)calloc(1, sizeof(struct E2SM_KPM_EventTriggerDefinition_Format1__policyTest_List));
	innerDef->policyTest_List = policyTestList;
	
	int index = 0;
	while(index < event_trigger_count) {
		Trigger_ConditionIE_Item_t *triggerCondition = (Trigger_ConditionIE_Item_t *)calloc(1, sizeof(Trigger_ConditionIE_Item_t));
		assert(triggerCondition != 0);
		triggerCondition->report_Period_IE = RT_periods[index];

		ASN_SEQUENCE_ADD(&policyTestList->list, triggerCondition);
		index++;
	}

	asn_enc_rval_t encode_result;
    encode_result = aper_encode_to_buffer(&asn_DEF_E2SM_KPM_EventTriggerDefinition, NULL, eventTriggerDef, buffer, buf_size);
    ASN_STRUCT_FREE(asn_DEF_E2SM_KPM_EventTriggerDefinition, eventTriggerDef);
    if(encode_result.encoded == -1) {
        fprintf(stderr, "Cannot encode %s: %s\n", encode_result.failed_type->name, strerror(errno));
        return -1;
    } else {
	    return encode_result.encoded;
	}
}

ssize_t e2sm_encode_ric_action_definition(void *buffer, size_t buf_size, long ric_style_type) {
	E2SM_KPM_ActionDefinition_t *actionDef = (E2SM_KPM_ActionDefinition_t *)calloc(1, sizeof(E2SM_KPM_ActionDefinition_t));
	if(!actionDef) {
		fprintf(stderr, "alloc RIC ActionDefinition failed\n");
		return -1;
	}

	actionDef->ric_Style_Type = ric_style_type;

	asn_enc_rval_t encode_result;
    encode_result = aper_encode_to_buffer(&asn_DEF_E2SM_KPM_ActionDefinition, NULL, actionDef, buffer, buf_size);
    ASN_STRUCT_FREE(asn_DEF_E2SM_KPM_ActionDefinition, actionDef);
	if(encode_result.encoded == -1) {
	    fprintf(stderr, "Cannot encode %s: %s\n", encode_result.failed_type->name, strerror(errno));
	    return -1;
	} else {
    	return encode_result.encoded;
    }
}

E2SM_KPM_IndicationHeader_t* e2sm_decode_ric_indication_header(void *buffer, size_t buf_size) {
	asn_dec_rval_t decode_result;
    E2SM_KPM_IndicationHeader_t *indHdr = 0;
    decode_result = aper_decode_complete(NULL, &asn_DEF_E2SM_KPM_IndicationHeader, (void **)&indHdr, buffer, buf_size);
    if(decode_result.code == RC_OK) {
        return indHdr;
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2SM_KPM_IndicationHeader, indHdr);
        return NULL;
    }
}

void e2sm_free_ric_indication_header(E2SM_KPM_IndicationHeader_t* indHdr) {
	ASN_STRUCT_FREE(asn_DEF_E2SM_KPM_IndicationHeader, indHdr);
}

E2SM_KPM_IndicationMessage_t* e2sm_decode_ric_indication_message(void *buffer, size_t buf_size) {
	asn_dec_rval_t decode_result;
    E2SM_KPM_IndicationMessage_t *indMsg = 0;
    decode_result = aper_decode_complete(NULL, &asn_DEF_E2SM_KPM_IndicationMessage, (void **)&indMsg, buffer, buf_size);
    if(decode_result.code == RC_OK) {
    	return indMsg;
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2SM_KPM_IndicationMessage, indMsg);
        return NULL;
    }
}

void e2sm_free_ric_indication_message(E2SM_KPM_IndicationMessage_t* indMsg) {
	ASN_STRUCT_FREE(asn_DEF_E2SM_KPM_IndicationMessage, indMsg);
}

// added by sww, ITRI (BEGIN)
E2SM_KPM_IndicationHeader_Format1_t* e2sm_get_indication_header_format1(E2SM_KPM_IndicationHeader_t *indHdr) {
    return indHdr->choice.indicationHeader_Format1;
}

E2SM_KPM_IndicationMessage_Format1_t* e2sm_get_indication_message_format1(E2SM_KPM_IndicationMessage_t *indMsg) {
    return indMsg->indicationMessage.choice.indicationMessage_Format1;
}

GlobalKPMnode_gNB_ID_t *e2sm_get_global_kpm_node_gnb_id(GlobalKPMnode_ID_t *globalKpmNodeId) {
    return globalKpmNodeId->choice.gNB;
}

GlobalKPMnode_en_gNB_ID_t *e2sm_get_global_kpm_node_en_gnb_id(GlobalKPMnode_ID_t *globalKpmNodeId) {
    return globalKpmNodeId->choice.en_gNB;
}

GlobalKPMnode_ng_eNB_ID_t *e2sm_get_global_kpm_node_ng_enb_id(GlobalKPMnode_ID_t *globalKpmNodeId) {
    return globalKpmNodeId->choice.ng_eNB;
}

GlobalKPMnode_eNB_ID_t *e2sm_get_global_kpm_node_enb_id(GlobalKPMnode_ID_t *globalKpmNodeId) {
    return globalKpmNodeId->choice.eNB;
}

BIT_STRING_t e2sm_get_gnb_id_choice(GNB_ID_Choice_t gnbIdChoice) {
    return gnbIdChoice.choice.gnb_ID;
}

BIT_STRING_t e2sm_get_gnb_id(ENGNB_ID_t enGnbId) {
    return enGnbId.choice.gNB_ID;
}


BIT_STRING_t e2sm_get_enb_id_macro(ENB_ID_Choice_t enbIdChoice) {
    return enbIdChoice.choice.enb_ID_macro;
}

BIT_STRING_t e2sm_get_enb_id_short_macro(ENB_ID_Choice_t enbIdChoice) {
    return enbIdChoice.choice.enb_ID_shortmacro;
}

BIT_STRING_t e2sm_get_enb_id_long_macro(ENB_ID_Choice_t enbIdChoice) {
    return enbIdChoice.choice.enb_ID_longmacro;
}


BIT_STRING_t e2sm_get_macro_enb_id(ENB_ID_t enbId) {
    return enbId.choice.macro_eNB_ID;
}

BIT_STRING_t e2sm_get_home_enb_id(ENB_ID_t enbId) {
    return enbId.choice.home_eNB_ID;
}

BIT_STRING_t e2sm_get_short_macro_enb_id(ENB_ID_t enbId) {
    return enbId.choice.short_Macro_eNB_ID;
}

BIT_STRING_t e2sm_get_long_macro_enb_id(ENB_ID_t enbId) {
    return enbId.choice.long_Macro_eNB_ID;
}


GNB_DU_Name_t e2sm_get_gnb_du_name(GNB_Name_t *gnbName) {
    return gnbName->choice.gNB_DU_Name;
}

GNB_CU_CP_Name_t e2sm_get_gnb_cucp_name(GNB_Name_t *gnbName) {
    return gnbName->choice.gNB_CU_UP_Name;
}

GNB_CU_UP_Name_t e2sm_get_gnb_cuup_name(GNB_Name_t *gnbName) {
    return gnbName->choice.gNB_CU_UP_Name;
}

PM_Containers_List_t *e2sm_get_pm_container_list(E2SM_KPM_IndicationMessage_Format1_t *indMsg, int i) {
    return indMsg->pm_Containers.list.array[i];
}

ODU_PF_Container_t* e2sm_get_du_pf_container(PM_Containers_List_t *pmContainersList) {
    if (NULL == pmContainersList->performanceContainer) {
        return NULL;
    }
    return pmContainersList->performanceContainer->choice.oDU;
}

OCUCP_PF_Container_t* e2sm_get_cucp_pf_container(PM_Containers_List_t *pmContainersList) {
    if (NULL == pmContainersList->performanceContainer) {
        return NULL;
    }
    return pmContainersList->performanceContainer->choice.oCU_CP;
}

OCUUP_PF_Container_t* e2sm_get_cuup_pf_container(PM_Containers_List_t *pmContainersList) {
    if (NULL == pmContainersList->performanceContainer) {
        return NULL;
    }
    return pmContainersList->performanceContainer->choice.oCU_UP;
}

DU_Usage_Report_Per_UE_t* e2sm_get_du_usage_report_per_ue(PM_Containers_List_t *pmContainersList) {
    if (NULL == pmContainersList->theRANContainer) {
        return NULL;
    }
    return pmContainersList->theRANContainer->reportContainer.choice.oDU_UE;
}

CU_CP_Usage_Report_Per_UE_t* e2sm_get_cucp_usage_report_per_ue(PM_Containers_List_t *pmContainersList) {
    if (NULL == pmContainersList->theRANContainer) {
        return NULL;
    }
    return pmContainersList->theRANContainer->reportContainer.choice.oCU_CP_UE;
}

CU_UP_Usage_Report_Per_UE_t* e2sm_get_cuup_usage_report_per_ue(PM_Containers_List_t *pmContainersList) {
    if (NULL == pmContainersList->theRANContainer) {
        return NULL;
    }
    return pmContainersList->theRANContainer->reportContainer.choice.oCU_UP_UE;
}


CellResourceReportListItem_t *e2sm_get_cell_resource_report_list_item(ODU_PF_Container_t *pfContainer, int i) {
    return pfContainer->cellResourceReportList.list.array[i];
}

ServedPlmnPerCellListItem_t *e2sm_get_served_plmn_per_cell_list(CellResourceReportListItem_t *cellResourceReportListItem, int i) {
    return cellResourceReportListItem->servedPlmnPerCellList.list.array[i];
}


DU_Usage_Report_CellResourceReportItem_t *e2sm_get_du_usage_report_cell_resource_report_item(DU_Usage_Report_Per_UE_t *usageReportPerUe, int i) {
    return usageReportPerUe->cellResourceReportList.list.array[i];
}

DU_Usage_Report_UeResourceReportItem_t *e2sm_get_du_usage_report_ue_resource_report_item(DU_Usage_Report_CellResourceReportItem_t *cellResourceReportItem, int i) {
    return cellResourceReportItem->ueResourceReportList.list.array[i];
}

CU_CP_Usage_Report_CellResourceReportItem_t *e2sm_get_cucp_usage_report_cell_resource_report_item(CU_CP_Usage_Report_Per_UE_t *usageReportPerUe, int i) {
    return usageReportPerUe->cellResourceReportList.list.array[i];
}

CU_CP_Usage_Report_UeResourceReportItem_t *e2sm_get_cucp_usage_report_ue_resource_report_item(CU_CP_Usage_Report_CellResourceReportItem_t *cellResourceReportItem, int i) {
    return cellResourceReportItem->ueResourceReportList.list.array[i];
}

CU_UP_Usage_Report_CellResourceReportItem_t *e2sm_get_cuup_usage_report_cell_resource_report_item(CU_UP_Usage_Report_Per_UE_t *usageReportPerUe, int i) {
    return usageReportPerUe->cellResourceReportList.list.array[i];
}

CU_UP_Usage_Report_UeResourceReportItem_t *e2sm_get_cuup_usage_report_ue_resource_report_item(CU_UP_Usage_Report_CellResourceReportItem_t *cellResourceReportItem, int i) {
    return cellResourceReportItem->ueResourceReportList.list.array[i];
}

SliceToReportListItem_t *e2sm_get_slice_to_report_list_item(FGC_CUUP_PM_Format_t *pmFormat, int i) {
    return pmFormat->sliceToReportList.list.array[i];
}

PerQCIReportListItemFormat_t *e2sm_get_per_qci_report_list_item_format(EPC_CUUP_PM_Format_t *pmFormat, int i) {
    return pmFormat->perQCIReportList.list.array[i];
}

SlicePerPlmnPerCellListItem_t *e2sm_get_slice_per_cell_list_item(FGC_DU_PM_Container_t *pmContainer, int i) {
    return pmContainer->slicePerPlmnPerCellList.list.array[i];
}

PerQCIReportListItem_t *e2sm_get_per_qci_report_list_item(EPC_DU_PM_Container_t *pmContainer, int i) {
    return pmContainer->perQCIReportList.list.array[i];
}

PF_ContainerListItem_t *e2sm_get_pf_container_list_item(OCUUP_PF_Container_t *pfContainer, int i) {
    return pfContainer->pf_ContainerList.list.array[i];
}

FQIPERSlicesPerPlmnPerCellListItem_t *e2sm_get_fqi_per_slices_per_cell_list_item(SlicePerPlmnPerCellListItem_t *slicePerPlmnPerCellListItem, int i) {
    return slicePerPlmnPerCellListItem->fQIPERSlicesPerPlmnPerCellList.list.array[i];
}

FQIPERSlicesPerPlmnListItem_t *e2sm_get_fqi_per_slices_per_plmn_list_item(SliceToReportListItem_t *sliceToReportListItem, int i) {
    return sliceToReportListItem->fQIPERSlicesPerPlmnList.list.array[i];
}

PlmnID_List_t *e2sm_get_plmnid_list(PF_ContainerListItem_t *pfContainerListItem, int i) {
    return pfContainerListItem->o_CU_UP_PM_Container.plmnList.list.array[i];
}
// added by sww, ITRI (END)

