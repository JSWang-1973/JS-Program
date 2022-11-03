#ifndef	_WRAPPER_H_
#define	_WRAPPER_H_

#include "E2SM-KPM-EventTriggerDefinition.h"
#include "E2SM-KPM-EventTriggerDefinition-Format1.h"
#include "Trigger-ConditionIE-Item.h"
#include "E2SM-KPM-ActionDefinition.h"
#include "E2SM-KPM-IndicationHeader.h"
#include "E2SM-KPM-IndicationHeader-Format1.h"
#include "GlobalKPMnode-ID.h"
#include "GlobalKPMnode-gNB-ID.h"
#include "GlobalKPMnode-en-gNB-ID.h"
#include "GlobalKPMnode-ng-eNB-ID.h"
#include "GlobalKPMnode-eNB-ID.h"
#include "PLMN-Identity.h"
#include "GNB-ID-Choice.h"
#include "GNB-CU-UP-ID.h"
#include "GNB-DU-ID.h"
#include "ENGNB-ID.h"
#include "ENB-ID-Choice.h"
#include "ENB-ID.h"
#include "NRCGI.h"
#include "SNSSAI.h"
#include "GNB-Name.h"
#include "E2SM-KPM-IndicationMessage.h"
#include "E2SM-KPM-IndicationMessage-Format1.h"
#include "PM-Containers-List.h"
#include "PF-Container.h"
#include "RAN-Container.h"
#include "ODU-PF-Container.h"
#include "CellResourceReportListItem.h"
#include "ServedPlmnPerCellListItem.h"
#include "FGC-DU-PM-Container.h"
#include "EPC-DU-PM-Container.h"
#include "SlicePerPlmnPerCellListItem.h"
#include "FQIPERSlicesPerPlmnPerCellListItem.h"
#include "PerQCIReportListItem.h"
#include "OCUCP-PF-Container.h"
#include "OCUUP-PF-Container.h"
#include "PF-ContainerListItem.h"
#include "PlmnID-List.h"
#include "FGC-CUUP-PM-Format.h"
#include "SliceToReportListItem.h"
#include "FQIPERSlicesPerPlmnListItem.h"
#include "EPC-CUUP-PM-Format.h"
#include "PerQCIReportListItemFormat.h"
#include "DU-Usage-Report-Per-UE.h"
#include "DU-Usage-Report-CellResourceReportItem.h"
#include "DU-Usage-Report-UeResourceReportItem.h"
#include "CU-CP-Usage-Report-Per-UE.h"
#include "CU-CP-Usage-Report-CellResourceReportItem.h"
#include "CU-CP-Usage-Report-UeResourceReportItem.h"
#include "CU-UP-Usage-Report-Per-UE.h"
#include "CU-UP-Usage-Report-CellResourceReportItem.h"
#include "CU-UP-Usage-Report-UeResourceReportItem.h"

ssize_t e2sm_encode_ric_event_trigger_definition(void *buffer, size_t buf_size, size_t event_trigger_count, long *RT_periods);
ssize_t e2sm_encode_ric_action_definition(void *buffer, size_t buf_size, long ric_style_type);
E2SM_KPM_IndicationHeader_t* e2sm_decode_ric_indication_header(void *buffer, size_t buf_size);
void e2sm_free_ric_indication_header(E2SM_KPM_IndicationHeader_t* indHdr);
E2SM_KPM_IndicationMessage_t* e2sm_decode_ric_indication_message(void *buffer, size_t buf_size);
void e2sm_free_ric_indication_message(E2SM_KPM_IndicationMessage_t* indMsg);

// added by sww, ITRI (BEGIN)
E2SM_KPM_IndicationHeader_Format1_t* e2sm_get_indication_header_format1(E2SM_KPM_IndicationHeader_t *indHdr);
E2SM_KPM_IndicationMessage_Format1_t* e2sm_get_indication_message_format1(E2SM_KPM_IndicationMessage_t *inMsg);
PM_Containers_List_t *e2sm_get_pm_container_list(E2SM_KPM_IndicationMessage_Format1_t *indMsg, int i);
ODU_PF_Container_t* e2sm_get_du_pf_container(PM_Containers_List_t *pmContainersList);
OCUCP_PF_Container_t* e2sm_get_cucp_pf_container(PM_Containers_List_t *pmContainersList);
OCUUP_PF_Container_t* e2sm_get_cuup_pf_container(PM_Containers_List_t *pmContainersList);
DU_Usage_Report_Per_UE_t* e2sm_get_du_usage_report_per_ue(PM_Containers_List_t *pmContainersList);
CU_CP_Usage_Report_Per_UE_t* e2sm_get_cucp_usage_report_per_ue(PM_Containers_List_t *pmContainersList);
CU_UP_Usage_Report_Per_UE_t* e2sm_get_cuup_usage_report_per_ue(PM_Containers_List_t *pmContainersList);
CellResourceReportListItem_t *e2sm_get_cell_resource_report_list_item(ODU_PF_Container_t *pfContainer, int i);
ServedPlmnPerCellListItem_t *e2sm_get_served_plmn_per_cell_list(CellResourceReportListItem_t *cellResourceReportListItem, int i);

DU_Usage_Report_CellResourceReportItem_t *e2sm_get_du_usage_report_cell_resource_report_item(DU_Usage_Report_Per_UE_t *usageReportPerUe, int i);
DU_Usage_Report_UeResourceReportItem_t *e2sm_get_du_usage_report_ue_resource_report_item(DU_Usage_Report_CellResourceReportItem_t *cellResourceReportItem, int i);
CU_CP_Usage_Report_CellResourceReportItem_t *e2sm_get_cucp_usage_report_cell_resource_report_item(CU_CP_Usage_Report_Per_UE_t *usageReportPerUe, int i);
CU_CP_Usage_Report_UeResourceReportItem_t *e2sm_get_cucp_usage_report_ue_resource_report_item(CU_CP_Usage_Report_CellResourceReportItem_t *cellResourceReportItem, int i);
CU_UP_Usage_Report_CellResourceReportItem_t *e2sm_get_cuup_usage_report_cell_resource_report_item(CU_UP_Usage_Report_Per_UE_t *usageReportPerUe, int i);
CU_UP_Usage_Report_UeResourceReportItem_t *e2sm_get_cuup_usage_report_ue_resource_report_item(CU_UP_Usage_Report_CellResourceReportItem_t *cellResourceReportItem, int i);

SliceToReportListItem_t *e2sm_get_slice_to_report_list_item(FGC_CUUP_PM_Format_t *pmFormat, int i);
PerQCIReportListItemFormat_t *e2sm_get_per_qci_report_list_item_format(EPC_CUUP_PM_Format_t *pmFormat, int i);
SlicePerPlmnPerCellListItem_t *e2sm_get_slice_per_cell_list_item(FGC_DU_PM_Container_t *pmContainer, int i);
PerQCIReportListItem_t *e2sm_get_per_qci_report_list_item(EPC_DU_PM_Container_t *pmContainer, int i);
PF_ContainerListItem_t *e2sm_get_pf_container_list_item(OCUUP_PF_Container_t *pfContainer, int i);
FQIPERSlicesPerPlmnPerCellListItem_t *e2sm_get_fqi_per_slices_per_cell_list_item(SlicePerPlmnPerCellListItem_t *slicePerPlmnPerCellListItem, int i);
FQIPERSlicesPerPlmnListItem_t *e2sm_get_fqi_per_slices_per_plmn_list_item(SliceToReportListItem_t *sliceToReportListItem, int i);
PlmnID_List_t *e2sm_get_plmnid_list(PF_ContainerListItem_t *pfContainerListItem, int i);

GlobalKPMnode_gNB_ID_t *e2sm_get_global_kpm_node_gnb_id(GlobalKPMnode_ID_t *globalKpmNodeId);
GlobalKPMnode_en_gNB_ID_t *e2sm_get_global_kpm_node_en_gnb_id(GlobalKPMnode_ID_t *globalKpmNodeId);
GlobalKPMnode_ng_eNB_ID_t *e2sm_get_global_kpm_node_ng_enb_id(GlobalKPMnode_ID_t *globalKpmNodeId);
GlobalKPMnode_eNB_ID_t *e2sm_get_global_kpm_node_enb_id(GlobalKPMnode_ID_t *globalKpmNodeId);
BIT_STRING_t e2sm_get_gnb_id_choice(GNB_ID_Choice_t gnbIdChoice);
BIT_STRING_t e2sm_get_gnb_id(ENGNB_ID_t enGnbId);

BIT_STRING_t e2sm_get_enb_id_macro(ENB_ID_Choice_t enbIdChoice);
BIT_STRING_t e2sm_get_enb_id_short_macro(ENB_ID_Choice_t enbIdChoice);
BIT_STRING_t e2sm_get_enb_id_long_macro(ENB_ID_Choice_t enbIdChoice);

BIT_STRING_t e2sm_get_macro_enb_id(ENB_ID_t enbId);
BIT_STRING_t e2sm_get_home_enb_id(ENB_ID_t enbId);
BIT_STRING_t e2sm_get_short_macro_enb_id(ENB_ID_t enbId);
BIT_STRING_t e2sm_get_long_macro_enb_id(ENB_ID_t enbId);

GNB_DU_Name_t e2sm_get_gnb_du_name(GNB_Name_t *gnbName);
GNB_CU_CP_Name_t e2sm_get_gnb_cucp_name(GNB_Name_t *gnbName);
GNB_CU_UP_Name_t e2sm_get_gnb_cuup_name(GNB_Name_t *gnbName);
// added by sww, ITRI (END)

#endif /* _WRAPPER_H_ */
