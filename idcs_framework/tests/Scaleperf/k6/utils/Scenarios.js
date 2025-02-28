let predefined_framework_scenarios = {

    // Specific number of VUs to complete fixed number of total iteration (irrespective of iterations per user)
    total_iteration:{
      executor: 'shared-iterations',
      vus: __ENV.vuCount,
      iterations: __ENV.totalIteration,
      maxDuration: '1h'
    },
  
    // Specific number of iterations for each VU
    per_vu_iteration: {
      executor: 'per-vu-iterations',
      vus: __ENV.vuCount,
      iterations: 1,
      maxDuration: '1h'
    },
  
    // Specific number of VU's for certain time
    time: {
      executor: 'constant-vus',
      vus: __ENV.vuCount,
      duration: __ENV.executionTime
    },
  
    // Ramping VUs - increment the VU's after certain time interval
    ramp_users:{
        executors: 'ramping-vus',
        startVUs: 0,
        stages: [
            // yet to be decided
        ],
        gracefulRampDown: '0s'
    },
  
    // Constant Arrival rate - number of requests per second can be defined.
    fixed_request_rate: {
      executor: 'constant-arrival-rate',
      duration: __ENV.executionTime,
      rate: __ENV.requestRate,
      timeUnit: '1s',
      preAllocatedVUs: 1,
      maxVUs: 50
    },
  
    // Ramping arrival rate  - number of requests per second can be incremented after certain time interval 
    ramp_rate: {
        executor: 'ramping-arrival-rate',
        startRate: __ENV.startRate,
        timeUnit: '1m',
        preAllocatedVUs: 1,
        maxVUs: 50,
        stages: [
            // yet to be decided
        ]
    }
  
  }

  export function get_predefined_framework_scenarios(){
    return predefined_framework_scenarios
  }