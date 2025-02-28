.. _logs_and_traces:

Go Logging and Tracing Guide for Developers
===========================================

Introduction
------------

Effective logging and tracing are critical for the observability and debugging of applications. This guide provides best practices for developers working with Go to enhance the traceability of errors, improve observability, and ensure that logs are informative and secure.

Log Errors Before Returning Them
--------------------------------

**Objective:** To capture the full trace of errors across code layers.

**Practice:**

- Always log an error before returning it.
- Include context about the error, such as the function name and any relevant variables.
- Use structured logging to make it easier to search and filter logs.

**Example:**

.. code-block:: golang

    func doSomething() error {
        err := someOperation()
        if err != nil {
            log.Error(err, "doSomething: someOperation failed.")
            return err
        }
        return nil
    }

Import and use logkeys constants while logging fields
-----------------------------------------------------

**Objective:** We want to ensure that we use consistent key names while logging same fields across all IDC services. This will help us in indexing logs better to improve searchability. 
For the same, we have centralized list of keys that can be used while logging across all services instead of hardcoding string keys in log statements individually by developers.

**Practice:**

- Import logkeys from log package available in IDC repo
- use the existing key if there is one already defined in the above list otherwise create a new one in the file and it to log the field.

**Example:**
if you want to log cloud account id, we have a key already defined in the list and can be used as logkeys.CloudAccountId. 
On the other hand, if you want to log "new field" and it doesn't exist in logkeys, we can add 
NewField = "newField"  in logkeys.go and then use logkeys.NewField in the log statement as below

.. code-block:: golang

    func (s* Service) DoAnotherThing(ctx context.Context, obj Params) {
        ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Service.DoAnotherThing").WithValues(logkeys.CloudAccountId, obj.GetId()).Start()
        defer span.End()
        // Your code here

        logger.Info("example of new key", logkeys.NewField, newFieldValue)
    }

Use obs.LogAndSpanFromContext or obs.LogAndSpanFromContextOrGlobal Instead of log.FromContext
---------------------------------------------------------------------------------------------

**Objective:** To enhance observability capabilities—such as real-time monitoring, alerting, and root cause analysis—adding default metadata to logs and spans is essential. This metadata should encompass not only basic information like service name, host, environment, log level, and timestamps for logs, but also trace IDs, span IDs, and parent IDs for spans. Additionally, it's important to include context information that may be relevant to the action or the environment, such as cloud account IDs, region, machine instance identifiers, or any other parameters that could be crucial for diagnosing issues. Spans are particularly necessary in distributed systems where a request traverses multiple services or components, as they enable the tracing of the request's path and performance across the entire system, providing a comprehensive view of the transaction flow.

**Practice:**

- Replace `log.FromContext` with `obs.LogAndSpanFromContext` or `obs.LogAndSpanFromContextOrGlobal` to enrich logs with tracing information.
- Ensure that the context passed to functions contains the necessary metadata for observability.
- When using obs.LogAndSpanFromContext to enrich logs with tracing information, be mindful of the span's lifecycle, especially for long-running methods. If a method's execution is expected to last for the duration of a server request or is particularly lengthy, avoid creating a span unless it provides critical insights or is essential for troubleshooting. Unnecessary spans can lead to trace bloat and may negatively impact the performance and clarity of the tracing data. Always consider the value and cost of each span before adding it to your context

**Example:**

.. code-block:: golang

    func (s* Service) DoAnotherThing(ctx context.Context, obj Params) {
        ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Service.DoAnotherThing").WithValues("cloudAccountId", obj.GetId()).Start()
        defer span.End()
        // Your code here
    }

Use Different Log Levels for Debug
----------------------------------

**Objective:** To provide more instruction to enable debug log mode and standardize it.

**Practice:**

- Use log levels appropriately.
- Enable DEBUG (or level "9") level logging through configuration or environment variables.
- Ensure DEBUG logs provide detailed information useful for troubleshooting.

**Example:**

.. code-block:: golang

    func fetchData() {
        _, logger, _ := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("fetchData").Start()

        logger.V(9).Info("this is a debug message")
    }

Have "BEGIN" and "END" Log Messages in Debug Level
--------------------------------------------------

**Objective:** To clearly mark the start and end of function executions.

**Practice:**

- Log a "BEGIN" message at the start of a function using the same verbosity level as the surrounding informational messages.
- Log an "END" message before exiting a function, matching the verbosity level of the "BEGIN" message and informational logs.
- Include function names and any relevant identifiers in the messages.
- For critical initialization functions, such as service startup, use Info without a verbosity level.

**Example:**

.. code-block:: golang

    func (s *Service) ProcessRequest(req *Request) {
        ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Service.ProcessRequest").WithValues("cloudAccountId", req.GetId()).Start()

        // If surrounding Info messages are at V(9), use V(9) for BEGIN and END
        logger.V(9).Info("BEGIN") // Adjust V(9) according to the verbosity level of surrounding Info messages
        defer logger.V(9).Info("END")
        // Processing code here
    }

Note: Ensure that the verbosity level for 'BEGIN' and 'END' messages (V(9) in the example) is consistent with the level applied to informational messages within the same function.If informational messages are at the default level V(0), then 'BEGIN' and 'END' should also be logged at V(0).Info to maintain consistency. For critical initialization messages, such as service startup, always use Info without a verbosity level to ensure visibility.

Add Relevant Data for Spans and Logs
------------------------------------

**Objective:** To ensure logs and spans contain necessary information without compromising PII.

**Practice:**

- Include relevant data such as identifiers, timestamps, and status codes in logs and spans.
- Avoid logging sensitive information (PII).
- Review logging practices during PR reviews as part of the definition of done.

**Example:**

.. code-block:: golang

    func doSomething(ctx context.Context, req Params) {
        logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("doSomething").WithValues("cloudAccountId", req.GetId()).Start()
        defer span.End()

        importantInfo := externalLibrary.doOtherThing(req)
        span.SetAttributes(attribute.String("importantAttribute", importantInfo.attribute))

        // ...
    }

Null Safe Logging
-----------------

**Objective:** To prevent logging nil values that could cause panics.

**Practice:**

- Use `.Get()` methods with null checks before logging values.
- Avoid direct logging of variables that could potentially be nil.

**Example:**

.. code-block:: golang

    func printUserDetails(log Log, user *User) {
        if user != nil {
            log.Info("User details: %s", user.GetDetails())
        } else {
            log.Info("User details are not available")
        }
    }

Single Function Call in Loops
-----------------------------

**Objective:** To isolate single actions per call within loops for better traceability.

**Practice:**

- Refactor loops to call a single function that encapsulates the action to be performed.
- This approach simplifies debugging and tracing by providing a clear start and end point for each iteration's action.

**Example:**

.. code-block:: golang

    for _, item := range items {
        processItem(item)
    }

    func processItem(item ItemType) {
        logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("processItem").WithValues("cloudAccountId", item.GetId()).Start()
        logger.V(9).Info("BEGIN")
        defer logger.V(9).Info("END")
        defer span.End()

        // Process code
    }

Single Values Per Log
---------------------

**Objective:** To avoid logging entire objects, which can be verbose and may include sensitive data.

**Practice:**

- Log individual fields or properties instead of entire objects.
- Ensure that the logged information is relevant and does not expose sensitive data.

**Example:**

.. code-block:: golang

    func logUserAction(log Log, payload Payload) {
        log.Error("user had error", "userId", payload.userID, "action" payload.action)
        // Avoid: log.Error("user had error", "payload", payload)
    }

Conclusion
----------

By following these logging and tracing best practices, developers can create a more maintainable, observable, and secure Go application. Remember to review and update logging strategies regularly to adapt to new requirements and to ensure compliance with data protection regulations
