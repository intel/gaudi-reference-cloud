package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

func init() {
	metrics.Registry.MustRegister(BMHControllerErrors)
	metrics.Registry.MustRegister(NetworkNodeControllerErrors)
	metrics.Registry.MustRegister(DeviceManagerErrors)
}

const (
	// metrics labels
	MetricsLabelApplication        = "application"
	MetricsLabelComponent          = "component"
	MetricsLabelErrorType          = "error_type"
	MetricsLabelNodeGroupName      = "node_group_name"
	MetricsLabelHostName           = "host_name"
	MetricsLabelSwitchFQDN         = "switch_fqdn"
	MetricsLabelSwitchPort         = "switch_port"
	MetricsLabelKubeconfigFilePath = "kubeconfig_file_path"
	// error types
	ErrorTypeMissingSwitchCR            = "MissingSwitchCR"
	ErrorTypePortAlreadyOwned           = "PortAlreadyOwned"
	ErrorTypeFailedToCreateSwitchClient = "FailedToCreateSwitchClient"
	ErrorTypeCreateBMHControllerFailed  = "CreateBMHControllerFailed"
	// const values
	LabelValueNone = "none"
)

// note: when designing the Prometheus metrics, remember to consider the cardinality issue, which means we should not put too many labels into one single metrics,
// as the number of combinations for one Prometheus series will become huge and lead to performance issue for Prometheus. Below is one of the exmaple that we should avoid to use.

// one global metrics, using lables to identify errrs
// var sdnErrors = prometheus.NewGaugeVec(
//
//	prometheus.GaugeOpts{
//		Name:        "sdn_errors",
//		Help:        "Will be set to 1 if a switch is expected by BMH but does not exist. 0 otherwise.",
//		ConstLabels: prometheus.Labels{}, // static labels
//	},
//	// labels for all the errors need to be added to this list,
//	// total possible series for this metrics will be (#component * #error * #host_name * #switch_fqdn * #switch_port)
//	[]string{"component", "error", "host_name", "switch_fqdn", "switch_port"},
//
// )

var (

	/*
		General Metrics - Add the metrics here if you think it doesn't belong to any subcomponents or it can ben shared across multiple components.
	*/

	/*
		BMHControllerManager Metrics
	*/
	BMHControllerManagerErrors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "bmh_controller_manager_error",
			Help:        "BMH Controller Manager errors",
			ConstLabels: prometheus.Labels{MetricsLabelApplication: "sdn-controller", MetricsLabelComponent: "bmh-controller-manager"},
		},
		[]string{MetricsLabelErrorType, MetricsLabelKubeconfigFilePath},
	)

	/*
		BMHController Metrics
	*/
	// name the metrics with the prefix of the subcomponent name.
	BMHControllerErrors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "bmh_controller_error",
			Help:        "BMH Controller errors",
			ConstLabels: prometheus.Labels{MetricsLabelApplication: "sdn-controller", MetricsLabelComponent: "bmh-controller"},
		},
		// add all the dynamic labels that we may use here.

		// one of the downside of using one general metrics to handle multiple errors, is, we need to define all the labels we need for all the errors,
		// and when we are calling metrics.BMHControllerErrors.With(), we need to provide all the labels and values, otherwide it will panic.
		// in this case, we need to assgin the LabelValueNone to the empty labels.
		[]string{MetricsLabelErrorType, MetricsLabelHostName, MetricsLabelSwitchFQDN},
	)

	/*
		SwitchPortController Metrics
	*/
	SwitchPortControllerErrors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "switch_port_controller_error",
			Help:        "SwitchPort Controller errors",
			ConstLabels: prometheus.Labels{MetricsLabelApplication: "sdn-controller", MetricsLabelComponent: "switch-port-controller"},
		},
		[]string{MetricsLabelErrorType, MetricsLabelSwitchFQDN, MetricsLabelSwitchPort},
	)

	/*
		SwitchController Metrics
	*/
	SwitchControllerErrors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "switch_controller_error",
			Help:        "Switch Controller errors",
			ConstLabels: prometheus.Labels{MetricsLabelApplication: "sdn-controller", MetricsLabelComponent: "switch-controller"},
		},
		[]string{MetricsLabelErrorType, MetricsLabelSwitchFQDN},
	)

	/*
		NetworkNodeController Metrics
	*/
	NetworkNodeControllerErrors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "network_node_controller_error",
			Help:        "NetworkNode Controller errors",
			ConstLabels: prometheus.Labels{MetricsLabelApplication: "sdn-controller", MetricsLabelComponent: "network-node-controller"},
		},
		[]string{MetricsLabelErrorType, MetricsLabelHostName},
	)

	/*
		NodeGroupController Metrics
	*/
	NodeGroupControllerErrors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "node_group_controller_error",
			Help:        "NodeGroup Controller errors",
			ConstLabels: prometheus.Labels{MetricsLabelApplication: "sdn-controller", MetricsLabelComponent: "node-group-controller"},
		},
		[]string{MetricsLabelErrorType, MetricsLabelHostName},
	)

	/*
	   Device manager Metrics
	*/
	DeviceManagerErrors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "device_manager_errors",
			Help:        "DeviceManager errors",
			ConstLabels: prometheus.Labels{MetricsLabelApplication: "sdn-controller", MetricsLabelComponent: "device-manager"},
		},
		[]string{MetricsLabelErrorType, MetricsLabelSwitchFQDN},
	)
)
