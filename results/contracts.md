# Vacuum Fork Contract Report

- upstream: upstream-github.com/daveshanley/vacuum
- fork:     fork-github.com/buraksekili/vacuum

The parity table keys packages by short name (e.g. `motor`), so upstream's `github.com/daveshanley/vacuum/motor` and fork's `github.com/buraksekili/vacuum/motor` are compared under the same row. The module-path rename itself is a consumer-visible change and factors into the `breaking-drift` verdict.

Verdict: **breaking-drift**

The fork has renamed, removed, or re-signed symbols that upstream callers use, OR the module path itself has changed. Consumers cannot swap the libraries without updating their import paths and/or call sites. See the parity table for specifics.

## Parity table

| Package | Kind | Symbol | Upstream | Fork | Status |
|---|---|---|---|---|---|
| model | func | BuildFunctionResult | func(key string, message string, value interface{}) RuleFunctionResult | func(key string, message string, value interface{}) RuleFunctionResult | match |
| model | func | BuildFunctionResultString | func(message string) RuleFunctionResult | func(message string) RuleFunctionResult | match |
| model | func | BuildFunctionResultWithDescription | func(desc string, key string, message string, value interface{}) RuleFunctionResult | func(desc string, key string, message string, value interface{}) RuleFunctionResult | match |
| model | func | BuildOperationFieldPath | func(path string, method string, field string) string |  | removed |
| model | func | BuildPooledFunctionResult | func(key string, message string, value interface{}) *RuleFunctionResult |  | removed |
| model | func | BuildPooledFunctionResultWithDescription | func(desc string, key string, message string, value interface{}) *RuleFunctionResult |  | removed |
| model | func | BuildResponsePath | func(path string, method string, code string) string |  | removed |
| model | func | CastToRuleAction | func(action interface{}) *RuleAction | func(action interface{}) *RuleAction | match |
| model | func | CompileRegex | func(context RuleFunctionContext, pattern string, results *[]RuleFunctionResult) *regexp.Regexp | func(context RuleFunctionContext, pattern string, results *[]RuleFunctionResult) *regexp.Regexp | match |
| model | func | FormatMatches | func(ruleFormat string, specFormat string) bool |  | removed |
| model | func | GetJSONPathBuilder | func() *JSONPathBuilder |  | removed |
| model | func | GetPooledRuleFunctionResult | func() *RuleFunctionResult |  | removed |
| model | func | GetStringTemplates | func() *StringTemplates |  | removed |
| model | func | MapPathAndNodesToResults | func(path string, startNode *yaml.Node, endNode *yaml.Node, results []RuleFunctionResult) []RuleFunctionResult | func(path string, startNode *yaml.Node, endNode *yaml.Node, results []RuleFunctionResult) []RuleFunctionResult | match |
| model | func | NewRuleFunctionResultFromPool | func(message string, path string, ruleId string, ruleSeverity string) *RuleFunctionResult |  | removed |
| model | func | NewRuleResultSet | func(results []RuleFunctionResult) *RuleResultSet | func(results []RuleFunctionResult) *RuleResultSet | match |
| model | func | NewRuleResultSetPointer | func(results []*RuleFunctionResult) *RuleResultSet | func(results []*RuleFunctionResult) *RuleResultSet | match |
| model | func | NewStringTemplates | func() *StringTemplates |  | removed |
| model | func | ReturnPooledRuleFunctionResult | func(result *RuleFunctionResult) |  | removed |
| model | func | ValidateRuleFunctionContextAgainstSchema | func(ruleFunction RuleFunction, ctx RuleFunctionContext) (bool, []string) | func(ruleFunction RuleFunction, ctx RuleFunctionContext) (bool, []string) | match |
| model | type | AutoFixFunction | named(0 fields) |  | removed |
| model | type | Example | struct(4 fields) | struct(4 fields) | match |
| model | type | IgnoredItems | named(0 fields) | named(0 fields) | match |
| model | type | JSONPathBuilder | struct(0 fields) |  | removed |
| model | type | Rule | struct(16 fields) | struct(15 fields) | signature-changed |
| model | type | RuleAction | struct(3 fields) | struct(3 fields) | match |
| model | type | RuleCategory | struct(3 fields) | struct(3 fields) | match |
| model | type | RuleCategoryResult | struct(9 fields) | struct(9 fields) | match |
| model | type | RuleFunction | interface(3 fields) | interface(3 fields) | match |
| model | type | RuleFunctionContext | struct(13 fields) | struct(9 fields) | signature-changed |
| model | type | RuleFunctionProperty | struct(2 fields) | struct(2 fields) | match |
| model | type | RuleFunctionResult | struct(13 fields) | struct(12 fields) | signature-changed |
| model | type | RuleFunctionSchema | struct(7 fields) | struct(7 fields) | match |
| model | type | RuleResultSet | struct(6 fields) | struct(4 fields) | signature-changed |
| model | type | RuleResultsForCategory | struct(2 fields) | struct(2 fields) | match |
| model | type | SearchResult | struct(3 fields) | struct(3 fields) | match |
| model | type | StringTemplates | struct(0 fields) |  | removed |
| model | method | JSONPathBuilder.Build | func() string |  | removed |
| model | method | JSONPathBuilder.Field | func(field string) *JSONPathBuilder |  | removed |
| model | method | JSONPathBuilder.Index | func(index int) *JSONPathBuilder |  | removed |
| model | method | JSONPathBuilder.Key | func(key string) *JSONPathBuilder |  | removed |
| model | method | JSONPathBuilder.Reset | func() *JSONPathBuilder |  | removed |
| model | method | JSONPathBuilder.Root | func() *JSONPathBuilder |  | removed |
| model | method | Rule.GetSeverityAsIntValue | func() int | func() int | match |
| model | method | Rule.ToJSON | func() string | func() string | match |
| model | method | RuleFunctionContext.ClearOptionsCache | func() |  | removed |
| model | method | RuleFunctionContext.GetOptionsStringMap | func() map[string]string |  | removed |
| model | method | RuleFunctionSchema.GetPropertyDescription | func(name string) string | func(name string) string | match |
| model | method | RuleResultSet.AddFixedResults | func(fixedResults []RuleFunctionResult) |  | removed |
| model | method | RuleResultSet.CalculateCategoryHealth | func(category string) int | func(category string) int | match |
| model | method | RuleResultSet.GenerateSpectralReport | func(source string) []reports.SpectralReport | func(source string) []reports.SpectralReport | match |
| model | method | RuleResultSet.GetErrorCount | func() int | func() int | match |
| model | method | RuleResultSet.GetErrorsByRuleCategory | func(category string) []*RuleFunctionResult | func(category string) []*RuleFunctionResult | match |
| model | method | RuleResultSet.GetHintByRuleCategory | func(category string) []*RuleFunctionResult | func(category string) []*RuleFunctionResult | match |
| model | method | RuleResultSet.GetHintCount | func() int |  | removed |
| model | method | RuleResultSet.GetInfoByRuleCategory | func(category string) []*RuleFunctionResult | func(category string) []*RuleFunctionResult | match |
| model | method | RuleResultSet.GetInfoCount | func() int | func() int | match |
| model | method | RuleResultSet.GetResultsByRuleCategory | func(category string) []*RuleFunctionResult | func(category string) []*RuleFunctionResult | match |
| model | method | RuleResultSet.GetResultsForCategoryWithLimit | func(category string, limit int) *RuleResultsForCategory | func(category string, limit int) *RuleResultsForCategory | match |
| model | method | RuleResultSet.GetRuleResultsForCategory | func(category string) *RuleResultsForCategory | func(category string) *RuleResultsForCategory | match |
| model | method | RuleResultSet.GetWarnCount | func() int | func() int | match |
| model | method | RuleResultSet.GetWarningsByRuleCategory | func(category string) []*RuleFunctionResult | func(category string) []*RuleFunctionResult | match |
| model | method | RuleResultSet.Len | func() int | func() int | match |
| model | method | RuleResultSet.Less | func(i int, j int) bool | func(i int, j int) bool | match |
| model | method | RuleResultSet.PrepareForSerialization | func(info *datamodel.SpecInfo) | func(info *datamodel.SpecInfo) | match |
| model | method | RuleResultSet.ResetCategoryCache | func() |  | removed |
| model | method | RuleResultSet.ResetCounts | func() | func() | match |
| model | method | RuleResultSet.SortResultsByLineNumber | func() []*RuleFunctionResult | func() []*RuleFunctionResult | match |
| model | method | RuleResultSet.Swap | func(i int, j int) | func(i int, j int) | match |
| model | method | RuleResultsForCategory.Len | func() int | func() int | match |
| model | method | RuleResultsForCategory.Less | func(i int, j int) bool | func(i int, j int) bool | match |
| model | method | RuleResultsForCategory.Swap | func(i int, j int) | func(i int, j int) | match |
| model | method | StringTemplates.BuildAPIKeyMessage | func(key string) string |  | removed |
| model | method | StringTemplates.BuildAlphabeticalMessage | func(ruleMessage string, item1 string, item2 string) string |  | removed |
| model | method | StringTemplates.BuildArrayPath | func(base string, index int) string |  | removed |
| model | method | StringTemplates.BuildBothDefinedMessage | func(ruleMessage string, field1 string, field2 string) string |  | removed |
| model | method | StringTemplates.BuildCachedFieldValidationMessage | func(ruleMessage string, field string, condition string) string |  | removed |
| model | method | StringTemplates.BuildCachedPatternMessage | func(ruleMessage string, value string, pattern string) string |  | removed |
| model | method | StringTemplates.BuildCredentialsMessage | func(param string) string |  | removed |
| model | method | StringTemplates.BuildEnumValidationMessage | func(ruleMessage string, value string, allowedValues interface{}) string |  | removed |
| model | method | StringTemplates.BuildFieldMessage | func(ruleMessage string, field string, message string) string |  | removed |
| model | method | StringTemplates.BuildFieldMustNotMessage | func(ruleMessage string, field string, condition string) string |  | removed |
| model | method | StringTemplates.BuildFieldValidationMessage | func(ruleMessage string, field string, condition string) string |  | removed |
| model | method | StringTemplates.BuildHTTPVerbInPathMessage | func(path string, verb string) string |  | removed |
| model | method | StringTemplates.BuildJSONPath | func(base string, field string) string |  | removed |
| model | method | StringTemplates.BuildKebabCaseMessage | func(segments string) string |  | removed |
| model | method | StringTemplates.BuildMissingExampleMessage | func(propName string) string |  | removed |
| model | method | StringTemplates.BuildMissingRequiredMessage | func(itemType string, item string, context string) string |  | removed |
| model | method | StringTemplates.BuildNumericalOrderingMessage | func(ruleMessage string, value1 string, value2 string) string |  | removed |
| model | method | StringTemplates.BuildOWASPResponseMessage | func(code string, headers string) string |  | removed |
| model | method | StringTemplates.BuildPatternMatchMessage | func(ruleMessage string, pattern string) string |  | removed |
| model | method | StringTemplates.BuildPatternMessage | func(ruleMessage string, value string, pattern string) string |  | removed |
| model | method | StringTemplates.BuildPropertyArrayPath | func(base string, property string, index int) string |  | removed |
| model | method | StringTemplates.BuildQuotedPath | func(base string, field string) string |  | removed |
| model | method | StringTemplates.BuildRegexCompileErrorMessage | func(ruleMessage string, pattern string, errorMsg string) string |  | removed |
| model | method | StringTemplates.BuildRequiredFieldMessage | func(field string) string |  | removed |
| model | method | StringTemplates.BuildSecurityDefinedMessage | func(path string, method string) string |  | removed |
| model | method | StringTemplates.BuildSecurityEmptyMessage | func(path string, method string) string |  | removed |
| model | method | StringTemplates.BuildSecurityNullElementsMessage | func(path string, method string) string |  | removed |
| model | method | StringTemplates.BuildTypeErrorMessage | func(ruleMessage string, value string, typeDesc string, errorMessage string) string |  | removed |
| model | method | StringTemplates.BuildUnknownSchemaTypeMessage | func(schemaType string) string |  | removed |
| motor | func | ApplyRulesToRuleSet | func(execution *RuleSetExecution) *RuleSetExecutionResult | func(execution *RuleSetExecution) *RuleSetExecutionResult | match |
| motor | func | BuildRolodexFromIndexConfig | func(indexConfig *index.SpecIndexConfig, customFS fs.FS) (*index.Rolodex, error) | func(indexConfig *index.SpecIndexConfig, customFS fs.FS) (*index.Rolodex, error) | match |
| motor | func | CreateRuleComposer | func() *RuleComposer | func() *RuleComposer | match |
| motor | type | RuleComposer | struct(0 fields) | struct(0 fields) | match |
| motor | type | RuleSetExecution | struct(35 fields) | struct(26 fields) | signature-changed |
| motor | type | RuleSetExecutionResult | struct(11 fields) | struct(9 fields) | signature-changed |
| motor | method | RuleComposer.ComposeRuleSet | func(ruleset []byte) (*rulesets.RuleSet, error) | func(ruleset []byte) (*rulesets.RuleSet, error) | match |
| motor | method | RuleSetExecutionResult.Release | func() |  | removed |
| rulesets | func | BuildDefaultRuleSets | func() RuleSets | func() RuleSets | match |
| rulesets | func | BuildDefaultRuleSetsWithLogger | func(logger *slog.Logger) RuleSets | func(logger *slog.Logger) RuleSets | match |
| rulesets | func | CheckForLocalExtends | func(extends map[string]string) bool | func(extends map[string]string) bool | match |
| rulesets | func | CheckForRemoteExtends | func(extends map[string]string) bool | func(extends map[string]string) bool | match |
| rulesets | func | CreateRuleSetFromData | func(data []byte) (*RuleSet, error) | func(data []byte) (*RuleSet, error) | match |
| rulesets | func | CreateRuleSetFromRuleMap | func(rules map[string]*model.Rule) *RuleSet | func(rules map[string]*model.Rule) *RuleSet | match |
| rulesets | func | CreateRuleSetUsingJSON | func(jsonData []byte) (*RuleSet, error) | func(jsonData []byte) (*RuleSet, error) | match |
| rulesets | func | DownloadRemoteRuleSet | func(ctx context.Context, location string, httpClient *http.Client) (*RuleSet, error) | func(_ context.Context, location string) (*RuleSet, error) | signature-changed |
| rulesets | func | ExpandAliasReferences | func(aliases map[string][]string) (map[string][]string, error) |  | removed |
| rulesets | func | ExpandRuleGivenPaths | func(givenPaths []string, expandedAliases map[string][]string) ([]string, error) |  | removed |
| rulesets | func | FilterRulesForTurbo | func(rs *RuleSet) int |  | removed |
| rulesets | func | GenerateDefaultOpenAPIRuleSet | func() *RuleSet | func() *RuleSet | match |
| rulesets | func | GenerateOWASPOpenAPIRuleSet | func() *RuleSet | func() *RuleSet | match |
| rulesets | func | GetAPIServersRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetAllBuiltInRules | func() map[string]*model.Rule | func() map[string]*model.Rule | match |
| rulesets | func | GetAllOWASPRules | func() map[string]*model.Rule | func() map[string]*model.Rule | match |
| rulesets | func | GetAllOfConflictsRule | func() *model.Rule |  | removed |
| rulesets | func | GetCamelCasePropertiesRule | func() *model.Rule |  | removed |
| rulesets | func | GetComponentDescriptionsRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetContactPropertiesRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetDescriptionDuplicationRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetDuplicatePathsRule | func() *model.Rule |  | removed |
| rulesets | func | GetDuplicatedEntryInEnumRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetGlobalOperationTagsRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetInfoContactRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetInfoDescriptionRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetInfoLicenseRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetInfoLicenseSPDXRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetInfoLicenseUrlRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetMigrateZallyIgnoreRule | func() *model.Rule |  | removed |
| rulesets | func | GetMissingTypeRule | func() *model.Rule |  | removed |
| rulesets | func | GetNoEvalInMarkdownRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetNoRefSiblingsRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetNoRequestBodyRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetNoScriptTagsInMarkdownRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetNoVerbsInPathRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetNullableEnumRule | func() *model.Rule |  | removed |
| rulesets | func | GetOAS2APIHostRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOAS2APISchemesRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOAS2DiscriminatorRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOAS2FormDataConsumesRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOAS2HostNotExampleRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOAS2HostTrailingSlashRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOAS2ParameterDescriptionRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOAS2PolymorphicAnyOfRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOAS2PolymorphicOneOfRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOAS2SchemaRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOAS2SecurityDefinedRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOAS2UnusedComponentRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOAS3ExamplesExternalCheck | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOAS3ExamplesMissingRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOAS3ExamplesRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOAS3HostNotExampleRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOAS3HostTrailingSlashRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOAS3NoRefSiblingsRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOAS3ParameterDescriptionRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOAS3SchemaRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOAS3SecurityDefinedRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOAS3UnusedComponentRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOWASPArrayLimitRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOWASPAuthInsecureSchemesRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOWASPConstrainedAdditionalPropertiesRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOWASPDefineErrorResponses401Rule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOWASPDefineErrorResponses429Rule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOWASPDefineErrorResponses500Rule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOWASPDefineErrorValidationRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOWASPIntegerFormatRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOWASPIntegerLimitRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOWASPJWTBestPracticesRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOWASPNoAPIKeysInURLRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOWASPNoAdditionalPropertiesRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOWASPNoCredentialsInURLRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOWASPNoHttpBasicRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOWASPNoNumericIDsRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOWASPProtectionGlobalSafeRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOWASPProtectionGlobalUnsafeRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOWASPProtectionGlobalUnsafeStrictRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOWASPRateLimitRetryAfterRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOWASPRateLimitRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOWASPSecurityHostsHttpsOAS3Rule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOWASPStringLimitRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOWASPStringRestrictedRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOpenApiTagsAlphabeticalRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOpenApiTagsRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOperationDescriptionRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOperationErrorResponseRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOperationIdRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOperationIdUniqueRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOperationIdValidInUrlRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOperationParametersRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOperationSingleTagRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOperationSuccessResponseRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetOperationTagsRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetPathDeclarationsMustExistRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetPathItemReferencesRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetPathNoTrailingSlashRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetPathNotIncludeQueryRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetPathParamsRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetPathsKebabCaseRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetPathsSpecificityOrderRule | func() *model.Rule |  | removed |
| rulesets | func | GetPostSuccessResponseRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetRecommendedOWASPRules | func() map[string]*model.Rule | func() map[string]*model.Rule | match |
| rulesets | func | GetRequiredFieldsDefinedRule | func() *model.Rule |  | removed |
| rulesets | func | GetSchemaTypeCheckRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetTagDescriptionRequiredRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetTypedEnumRule | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | GetUnnecessaryCombinatorRule | func() *model.Rule |  | removed |
| rulesets | func | LoadLocalRuleSet | func(_ context.Context, location string) (*RuleSet, error) | func(_ context.Context, location string) (*RuleSet, error) | match |
| rulesets | func | NoAmbiguousPaths | func() *model.Rule | func() *model.Rule | match |
| rulesets | func | ParseAliases | func(raw map[string]interface{}) (map[string]*ParsedAlias, error) |  | removed |
| rulesets | func | ResolveAliasesForFormat | func(parsed map[string]*ParsedAlias, specFormat string) map[string][]string |  | removed |
| rulesets | func | SniffOutAllExternalRules | func(ctx context.Context, rsm *ruleSetsModel, location string, visited []string, rs *RuleSet, remote bool, httpClient *http.Client) | func(ctx context.Context, rsm *ruleSetsModel, location string, visited []string, rs *RuleSet, remote bool) | signature-changed |
| rulesets | type | AliasTarget | struct(2 fields) |  | removed |
| rulesets | type | ParsedAlias | struct(2 fields) |  | removed |
| rulesets | type | RuleSet | struct(8 fields) | struct(6 fields) | signature-changed |
| rulesets | type | RuleSets | interface(4 fields) | interface(3 fields) | signature-changed |
| rulesets | type | SimpleAlias | named(0 fields) |  | removed |
| rulesets | type | TargetedAlias | struct(2 fields) |  | removed |
| rulesets | method | RuleSet.GetExtendsValue | func() map[string]string | func() map[string]string | match |
