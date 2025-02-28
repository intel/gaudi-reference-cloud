# ProtoLoadBalancerPool

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Port** | Pointer to **int32** | The port to route traffic to each instance. | [optional] 
**Monitor** | Pointer to [**ProtoLoadBalancerMonitorType**](ProtoLoadBalancerMonitorType.md) |  | [optional] [default to TCP]
**LoadBalancingMode** | Pointer to [**ProtoLoadBalancingMode**](ProtoLoadBalancingMode.md) |  | [optional] [default to ROUND_ROBIN]
**InstanceSelectors** | Pointer to **map[string]string** | (Optional) Map of string keys and values that controls how the lb pool members are selected.  One of instances or instanceSelectors is valid. | [optional] 
**InstanceResourceIds** | Pointer to **[]string** | (Optional) Set of Instances to make up the members of the pool. One of instances or instanceSelectors is valid. | [optional] 

## Methods

### NewProtoLoadBalancerPool

`func NewProtoLoadBalancerPool() *ProtoLoadBalancerPool`

NewProtoLoadBalancerPool instantiates a new ProtoLoadBalancerPool object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoLoadBalancerPoolWithDefaults

`func NewProtoLoadBalancerPoolWithDefaults() *ProtoLoadBalancerPool`

NewProtoLoadBalancerPoolWithDefaults instantiates a new ProtoLoadBalancerPool object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPort

`func (o *ProtoLoadBalancerPool) GetPort() int32`

GetPort returns the Port field if non-nil, zero value otherwise.

### GetPortOk

`func (o *ProtoLoadBalancerPool) GetPortOk() (*int32, bool)`

GetPortOk returns a tuple with the Port field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPort

`func (o *ProtoLoadBalancerPool) SetPort(v int32)`

SetPort sets Port field to given value.

### HasPort

`func (o *ProtoLoadBalancerPool) HasPort() bool`

HasPort returns a boolean if a field has been set.

### GetMonitor

`func (o *ProtoLoadBalancerPool) GetMonitor() ProtoLoadBalancerMonitorType`

GetMonitor returns the Monitor field if non-nil, zero value otherwise.

### GetMonitorOk

`func (o *ProtoLoadBalancerPool) GetMonitorOk() (*ProtoLoadBalancerMonitorType, bool)`

GetMonitorOk returns a tuple with the Monitor field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMonitor

`func (o *ProtoLoadBalancerPool) SetMonitor(v ProtoLoadBalancerMonitorType)`

SetMonitor sets Monitor field to given value.

### HasMonitor

`func (o *ProtoLoadBalancerPool) HasMonitor() bool`

HasMonitor returns a boolean if a field has been set.

### GetLoadBalancingMode

`func (o *ProtoLoadBalancerPool) GetLoadBalancingMode() ProtoLoadBalancingMode`

GetLoadBalancingMode returns the LoadBalancingMode field if non-nil, zero value otherwise.

### GetLoadBalancingModeOk

`func (o *ProtoLoadBalancerPool) GetLoadBalancingModeOk() (*ProtoLoadBalancingMode, bool)`

GetLoadBalancingModeOk returns a tuple with the LoadBalancingMode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLoadBalancingMode

`func (o *ProtoLoadBalancerPool) SetLoadBalancingMode(v ProtoLoadBalancingMode)`

SetLoadBalancingMode sets LoadBalancingMode field to given value.

### HasLoadBalancingMode

`func (o *ProtoLoadBalancerPool) HasLoadBalancingMode() bool`

HasLoadBalancingMode returns a boolean if a field has been set.

### GetInstanceSelectors

`func (o *ProtoLoadBalancerPool) GetInstanceSelectors() map[string]string`

GetInstanceSelectors returns the InstanceSelectors field if non-nil, zero value otherwise.

### GetInstanceSelectorsOk

`func (o *ProtoLoadBalancerPool) GetInstanceSelectorsOk() (*map[string]string, bool)`

GetInstanceSelectorsOk returns a tuple with the InstanceSelectors field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstanceSelectors

`func (o *ProtoLoadBalancerPool) SetInstanceSelectors(v map[string]string)`

SetInstanceSelectors sets InstanceSelectors field to given value.

### HasInstanceSelectors

`func (o *ProtoLoadBalancerPool) HasInstanceSelectors() bool`

HasInstanceSelectors returns a boolean if a field has been set.

### GetInstanceResourceIds

`func (o *ProtoLoadBalancerPool) GetInstanceResourceIds() []string`

GetInstanceResourceIds returns the InstanceResourceIds field if non-nil, zero value otherwise.

### GetInstanceResourceIdsOk

`func (o *ProtoLoadBalancerPool) GetInstanceResourceIdsOk() (*[]string, bool)`

GetInstanceResourceIdsOk returns a tuple with the InstanceResourceIds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstanceResourceIds

`func (o *ProtoLoadBalancerPool) SetInstanceResourceIds(v []string)`

SetInstanceResourceIds sets InstanceResourceIds field to given value.

### HasInstanceResourceIds

`func (o *ProtoLoadBalancerPool) HasInstanceResourceIds() bool`

HasInstanceResourceIds returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


