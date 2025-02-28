// Available compute product matrix
// [user_type] [region] [filter]
export const idcProductMatrix = {
    'Standard': {
        'us-staging-1': {
            'total': 3,
            // 'category_released': 3,
            'type_vm': 1,
            'type_bm': 3,
            'proc_cpu': 2,
            'proc_gpu': 1,
            'proc_ai': 0,
            'recommended_core': 2,
            'recommended_gpu': 1,
            'recommended_hpc': 0,
            'recommended_ai': 0
        },
        'us-staging-2': {
            'total': 3,
            // 'category_released': 3,
            'type_vm': 1,
            'type_bm': 2,
            'proc_cpu': 2,
            'proc_gpu': 1,
            'proc_ai': 0,
            'recommended_core': 2,
            'recommended_gpu': 1,
            'recommended_hpc': 0,
            'recommended_ai': 0
        },
        'us-staging-3': {
            'total': 3,
            // 'category_released': 3,
            'type_vm': 1,
            'type_bm': 2,
            'proc_cpu': 2,
            'proc_gpu': 1,
            'proc_ai': 0,
            'recommended_core': 2,
            'recommended_gpu': 1,
            'recommended_hpc': 0,
            'recommended_ai': 0
        },
        'us-staging-4': {
            'total': 3,
            // 'category_released': 3,
            'type_vm': 1,
            'type_bm': 2,
            'proc_cpu': 2,
            'proc_gpu': 1,
            'proc_ai': 0,
            'recommended_core': 2,
            'recommended_gpu': 1,
            'recommended_hpc': 0,
            'recommended_ai': 0
        },
        'us-staging-minimum': {
            'total': 3,
            'type_core_compute': 0,
            'proc_cpu': 2,
            'proc_gpu': 1,
            'proc_hpc': 0,
            'proc_ai': 0
        },
        'us-region-1': {
            'total': 2,
            // 'category_released': 3,
            'type_vm': 1,
            'type_bm': 2,
            'proc_cpu': 0,
            'proc_gpu': 0,
            'proc_ai': 0,
            'recommended_core': 0,
            'recommended_gpu': 0,
            'recommended_hpc': 0,
            'recommended_ai': 0
        },
        'us-region-2': {
            'total': 2,
            // 'category_released': 2,
            'type_vm': 1,
            'type_bm': 1,
            'proc_cpu': 0,
            'proc_gpu': 0,
            'proc_ai': 0,
            'recommended_core': 0,
            'recommended_gpu': 0,
            'recommended_hpc': 0,
            'recommended_ai': 0
        },
        'us-prod-minimum': {
            'total': 2,
            'type_core_compute': 1,
            'proc_gpu': 0,
            'proc_hpc': 0,
            'proc_ai': 0
        }
    },
    'Premium': {
        'us-staging-1': {
            'total': 5,
            // 'category_released': 6,
            'type_vm': 1,
            'type_bm': 5,
            'proc_cpu': 2,
            'proc_gpu': 1,
            'proc_ai': 1,
            'recommended_core': 1,
            'recommended_gpu': 1,
            'recommended_hpc': 1,
            'recommended_ai': 1
        },
        'us-staging-2': {
            'total': 5,
            // 'category_released': 6,
            'type_vm': 1,
            'type_bm': 5,
            'proc_cpu': 2,
            'proc_gpu': 1,
            'proc_ai': 1,
            'recommended_core': 1,
            'recommended_gpu': 1,
            'recommended_hpc': 1,
            'recommended_ai': 1
        },
        'us-staging-3': {
            'total': 5,
            // 'category_released': 6,
            'type_vm': 1,
            'type_bm': 5,
            'proc_cpu': 2,
            'proc_gpu': 1,
            'proc_ai': 1,
            'recommended_core': 1,
            'recommended_gpu': 1,
            'recommended_hpc': 1,
            'recommended_ai': 1
        },
	'us-staging-4': {
            'total': 5,
            // 'category_released': 6,
            'type_vm': 1,
            'type_bm': 5,
            'proc_cpu': 2,
            'proc_gpu': 1,
            'proc_ai': 1,
            'recommended_core': 1,
            'recommended_gpu': 1,
            'recommended_hpc': 1,
            'recommended_ai': 1
        },
        'us-staging-minimum': {
            'total': 4,
            'type_core_compute': 2,
            'proc_gpu': 1,
            'proc_hpc': 1,
            'proc_ai': 1
        },
        'us-region-1': {
            'total': 2,
            // 'category_released': 5,
            'type_vm': 1,
            'type_bm': 1,
            'proc_cpu': 1,
            'proc_gpu': 1,
            'proc_ai': 0,
            'recommended_core': 1,
            'recommended_gpu': 1,
            'recommended_hpc': 0,
            'recommended_ai': 0
        },
        'us-region-2': {
            'total': 2,
            // 'category_released': 4,
            'type_vm': 1,
            'type_bm': 1,
            'proc_cpu': 1,
            'proc_gpu': 1,
            'proc_ai': 0,
            'recommended_core': 1,
            'recommended_gpu': 1,
            'recommended_hpc': 0,
            'recommended_ai': 0
        },
        'us-prod-minimum': {
            'total': 3,
            'type_core_compute': 2,
            'proc_gpu': 1,
            'proc_hpc': 0,
            'proc_ai': 0
        }
    },
    'Internal': {}
}

export const idcRegionsShorts = {
    'us-region-1':  'reg1',
    'us-region-2':  'reg2',
    'us-region-3':  'reg3',
    'us-region-4':  'reg4',
    'us-staging-1': 'stg1',
    'us-staging-2': 'stg2',
    'us-staging-3': 'stg3',
    'us-staging-4': 'stg4',
    'us-qa1-1':     'qa1'
}

export const idcUsersShorts = {
    'Standard': 'std',
    'Premium':  'prem',
    'Preview':  'pview',
    'Internal': 'int'
}
