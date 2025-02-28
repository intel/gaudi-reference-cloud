export const typeDef = `

type preRunStats {
  TotalSpecs: Int
  SpecsThatWillRun: Int
}

type suiteConfig {
  RandomSeed: Int
  RandomizeAllSpecs: Boolean
  FocusStrings: String
  SkipStrings: String
  FocusFiles: String
  SkipFiles: String
  LabelFilter: String
  FailOnPending: Boolean
  FailFast: Boolean
  FlakeAttempts: Int
  MustPassRepeatedly: Int
  DryRun: Boolean
  PollProgressAfter: Int
  PollProgressInterval: Int
  Timeout: Int
  EmitSpecProgress: Boolean
  OutputInterceptorMode: String
  SourceRoots: String
  GracePeriod: Int
  ParallelProcess: Int
  ParallelTotal: Int
  ParallelHost: String
}

type leafNodeLocation {
  FileName: String
  LineNumber: Int
}

type codeLocation {
  FileName: String
  LineNumber: Int
}

type timelineLocation {
  Order: Int
  Time: String
}

type specEvent {
  SpecEventType: String
  CodeLocation: codeLocation
  TimelineLocation: timelineLocation
  Message: String
  NodeType: String
}

type specReport {
  ContainerHierarchyTexts: String
  ContainerHierarchyLocations: String
  ContainerHierarchyLabels: String
  LeafNodeType: String
  LeafNodeLocation: leafNodeLocation
  LeafNodeLabels: String
  LeafNodeText: String
  State: String
  StartTime: String
  EndTime: String
  RunTime: Int
  ParallelProcess: Int
  NumAttempts: Int
  MaxFlakeAttempts: Int
  MaxMustPassRepeatedly: Int
  SpecEvents: [specEvent]
}

type ginkgoLog {
  SuitePath: String
  SuiteDescription: String
  SuiteLabels: [String]
  SuiteSucceeded: Boolean
  SuiteHasProgrammaticFocus: Boolean
  SpecialSuiteFailureReasons: String
  PreRunStats: preRunStats
  StartTime: String
  EndTime: String
  RunTime: Int
  SuiteConfig: suiteConfig
  SpecReports: [specReport]
}

`;