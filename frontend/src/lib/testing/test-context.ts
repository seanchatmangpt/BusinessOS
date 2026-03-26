/**
 * Test Context Helpers
 *
 * Provides debugging context for test failures across all test types.
 * Use these to generate helpful error messages with debugging steps.
 */

export interface TestContext {
  category: string;
  details: Record<string, unknown>;
  errorMessage: string;
  debugSteps: string[];
}

/**
 * Context for API/network failures
 */
export function apiFailure(opts: {
  endpoint: string;
  expectedStatus?: number;
  actualStatus?: number;
  reason?: string;
}): TestContext {
  const endpoint = opts.endpoint || 'unknown';
  const expectedStatus = opts.expectedStatus || 200;
  const actualStatus = opts.actualStatus || 'unknown';
  const reason = opts.reason || 'unknown';

  const errorMessage = `
API Request Failed

Endpoint: ${endpoint}
Expected Status: ${expectedStatus}
Actual Status: ${actualStatus}
Reason: ${reason}

Debugging Steps:
  1. Is localhost:8001 running? Run: curl http://localhost:8001/api/health
  2. Check endpoint exists: curl -v http://localhost:8001${endpoint}
  3. Check BusinessOS logs: docker logs businessos-backend
  4. Verify response format: curl http://localhost:8001${endpoint} | jq .
  5. Check for CORS issues: curl -i -H "Origin: http://localhost:5173" http://localhost:8001${endpoint}
`;

  return {
    category: 'api_failure',
    details: opts,
    errorMessage,
    debugSteps: [
      'Verify service is running',
      'Check endpoint URL',
      'Check response status',
      'Check response format',
      'Check CORS headers',
    ],
  };
}

/**
 * Context for timing/async failures
 */
export function timingFailure(opts: {
  operation?: string;
  timeoutMs?: number;
  expectedValue?: unknown;
}): TestContext {
  const operation = opts.operation || 'unknown';
  const timeoutMs = opts.timeoutMs || 5000;
  const expectedValue = JSON.stringify(opts.expectedValue);

  const errorMessage = `
Timing/Async Failure Detected

Operation: ${operation}
Timeout: ${timeoutMs}ms
Expected: ${expectedValue}

Debugging Steps:
  1. Run test in isolation: npm test -- path/to/file.test.ts -t "test name"
  2. Increase timeout: jest.setTimeout(10000) at top of test
  3. Use waitFor with longer timeout:
     await waitFor(() => expect(element).toBeInTheDocument(), {
       timeout: 10000
     })
  4. Check for missing await on Promises
  5. Verify async setup/teardown functions

If test passes in isolation but fails with others:
  → It's a timing issue, not logic
  → Check for test pollution (beforeEach/afterEach cleanup)
  → Verify mocks are reset between tests
`;

  return {
    category: 'timing_failure',
    details: opts,
    errorMessage,
    debugSteps: [
      'Run test in isolation',
      'Increase timeout',
      'Add explicit waits',
      'Check test cleanup',
    ],
  };
}

/**
 * Context for logic/assertion failures
 */
export function logicFailure(opts: {
  operation?: string;
  expected?: unknown;
  actual?: unknown;
}): TestContext {
  const operation = opts.operation || 'unknown';
  const expected = JSON.stringify(opts.expected);
  const actual = JSON.stringify(opts.actual);

  const errorMessage = `
Logic/Assertion Failure

Operation: ${operation}
Expected: ${expected}
Actual: ${actual}

Debugging Steps:
  1. Add console.log at each step to trace execution
  2. Check data types (typeof value)
  3. Verify expected vs actual structure (especially arrays, objects)
  4. Check for null/undefined vs 0/"" false
  5. Use debugger statement and run: npm test -- --inspect-brk
     Then open chrome://inspect in Chrome

Try this:
  // In test file:
  console.log('step 1:', value);
  console.log('step 2:', value); // Use debugger; to pause execution

  Then run:
  npm test -- --inspect-brk --runInBand path/to/file.test.ts

  In Chrome: DevTools → Sources → Step through code
`;

  return {
    category: 'logic_failure',
    details: opts,
    errorMessage,
    debugSteps: [
      'Add console.log at each step',
      'Check data types',
      'Verify structure (array vs object)',
      'Use debugger and Chrome DevTools',
    ],
  };
}

/**
 * Context for resource exhaustion failures
 */
export function resourceFailure(opts: {
  resourceType?: string;
  limit?: unknown;
  actual?: unknown;
}): TestContext {
  const resourceType = opts.resourceType || 'unknown';
  const limit = JSON.stringify(opts.limit);
  const actual = JSON.stringify(opts.actual);

  const errorMessage = `
Resource Exhaustion Detected

Resource: ${resourceType}
Limit: ${limit}
Actual: ${actual}

Debugging Steps:
  1. Check memory usage: console.log(performance.memory)
  2. Monitor during test: watch -n 1 'ps aux | grep node'
  3. Check for unclosed resources (event listeners, timers)
  4. Verify cleanup in afterEach:
     afterEach(() => {
       jest.clearAllMocks();
       jest.clearAllTimers();
     });
  5. Use Chrome DevTools Memory profiler:
     npm test -- --inspect-brk
     Then: chrome://inspect → Memory → Heap snapshots

Common causes:
  • Event listeners not removed (addEventListener without removeEventListener)
  • Timers not cleared (setInterval without clearInterval)
  • Mocks not reset (jest.clearAllMocks() in afterEach)
  • Large data structures not freed
  • Unbounded array/map growth
`;

  return {
    category: 'resource_failure',
    details: opts,
    errorMessage,
    debugSteps: [
      'Check memory usage',
      'Verify cleanup in afterEach',
      'Check for leaked event listeners',
      'Check for leaked timers',
    ],
  };
}

/**
 * Context for flaky test failures
 */
export function flakyTest(opts: {
  testName?: string;
  passRate?: string;
  failurePattern?: string;
}): TestContext {
  const testName = opts.testName || 'unknown';
  const passRate = opts.passRate || 'unknown';
  const failurePattern = opts.failurePattern || 'unknown';

  const errorMessage = `
Flaky Test Detected

Test: ${testName}
Pass Rate: ${passRate}
Failure Pattern: ${failurePattern}

Debugging Steps:
  1. Run test multiple times:
     for i in {1..20}; do npm test -- path/to/file.test.ts; done
  2. Check for race conditions (async operations)
  3. Run serially (not in parallel):
     npm test -- --runInBand path/to/file.test.ts
  4. Check for shared state between tests
  5. Verify mocks are reset in beforeEach/afterEach

If fails only in parallel:
  → Missing wait (await, waitFor, etc.)
  → Or check for test isolation (one test affects another)

If fails unpredictably:
  → Likely timing or external service issue
  → Mock external services with fixed responses
  → Use fake time/clock (jest.useFakeTimers()) instead of real time
`;

  return {
    category: 'flaky_test',
    details: opts,
    errorMessage,
    debugSteps: [
      'Run test multiple times',
      'Run tests serially',
      'Check test isolation',
      'Mock external dependencies',
    ],
  };
}

/**
 * Helper to throw context as error
 */
export function throwWithContext(context: TestContext): never {
  throw new Error(context.errorMessage);
}

/**
 * Helper to log context and debug steps
 */
export function logContext(context: TestContext): void {
  console.error(context.errorMessage);
  console.group('Debug Steps');
  context.debugSteps.forEach((step, i) => {
    console.log(`${i + 1}. ${step}`);
  });
  console.groupEnd();
}

/**
 * Assertion helper with context
 */
export function assertWithContext(
  condition: boolean,
  message: string,
  context: TestContext,
): void {
  if (!condition) {
    console.error(`Assertion Failed: ${message}`);
    logContext(context);
    throw new Error(`${message}\n${context.errorMessage}`);
  }
}
