// ComplianceForge API Client
// Connects the Next.js frontend to the Golang backend

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

type RequestOptions = {
  method?: string;
  body?: unknown;
  headers?: Record<string, string>;
};

class ApiClient {
  private token: string | null = null;

  setToken(token: string) {
    this.token = token;
    if (typeof window !== 'undefined') {
      localStorage.setItem('cf_token', token);
    }
  }

  getToken(): string | null {
    if (!this.token && typeof window !== 'undefined') {
      this.token = localStorage.getItem('cf_token');
    }
    return this.token;
  }

  clearToken() {
    this.token = null;
    if (typeof window !== 'undefined') {
      localStorage.removeItem('cf_token');
    }
  }

  private async request<T>(path: string, options: RequestOptions = {}): Promise<T> {
    const { method = 'GET', body, headers = {} } = options;

    const requestHeaders: Record<string, string> = {
      'Content-Type': 'application/json',
      ...headers,
    };

    const token = this.getToken();
    if (token) {
      requestHeaders['Authorization'] = `Bearer ${token}`;
    }

    const response = await fetch(`${API_BASE}${path}`, {
      method,
      headers: requestHeaders,
      body: body ? JSON.stringify(body) : undefined,
    });

    if (response.status === 401) {
      this.clearToken();
      if (typeof window !== 'undefined') {
        window.location.href = '/auth/login';
      }
      throw new Error('Unauthorized');
    }

    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.error?.message || 'Request failed');
    }

    return data;
  }

  // ── Auth ──────────────────────────────────────────
  async login(email: string, password: string) {
    const data = await this.request<any>('/auth/login', {
      method: 'POST',
      body: { email, password },
    });
    if (data.data?.access_token) {
      this.setToken(data.data.access_token);
    }
    return data;
  }

  async me() {
    return this.request<any>('/auth/me');
  }

  // ── Dashboard ─────────────────────────────────────
  async getDashboard() {
    return this.request<any>('/dashboard/summary');
  }

  // ── Frameworks ────────────────────────────────────
  async getFrameworks() {
    return this.request<any>('/frameworks');
  }

  async getFramework(id: string) {
    return this.request<any>(`/frameworks/${id}`);
  }

  async getFrameworkControls(id: string, page = 1, pageSize = 20) {
    return this.request<any>(`/frameworks/${id}/controls?page=${page}&page_size=${pageSize}`);
  }

  async searchControls(query: string, limit = 20) {
    return this.request<any>(`/frameworks/controls/search?q=${encodeURIComponent(query)}&limit=${limit}`);
  }

  // ── Compliance ────────────────────────────────────
  async getComplianceScores() {
    return this.request<any>('/compliance/scores');
  }

  async getGapAnalysis(frameworkId?: string) {
    const params = frameworkId ? `?framework_id=${frameworkId}` : '';
    return this.request<any>(`/compliance/gaps${params}`);
  }

  async getCrossMapping() {
    return this.request<any>('/compliance/cross-mapping');
  }

  // ── Risks ─────────────────────────────────────────
  async getRisks(page = 1, pageSize = 20, sortBy = 'residual_risk_score', sortDir = 'desc') {
    return this.request<any>(`/risks?page=${page}&page_size=${pageSize}&sort_by=${sortBy}&sort_dir=${sortDir}`);
  }

  async getRisk(id: string) {
    return this.request<any>(`/risks/${id}`);
  }

  async createRisk(data: any) {
    return this.request<any>('/risks', { method: 'POST', body: data });
  }

  async getRiskHeatmap() {
    return this.request<any>('/risks/heatmap');
  }

  // ── Policies ──────────────────────────────────────
  async getPolicies(page = 1, pageSize = 20) {
    return this.request<any>(`/policies?page=${page}&page_size=${pageSize}`);
  }

  async getPolicy(id: string) {
    return this.request<any>(`/policies/${id}`);
  }

  async createPolicy(data: any) {
    return this.request<any>('/policies', { method: 'POST', body: data });
  }

  async publishPolicy(id: string) {
    return this.request<any>(`/policies/${id}/publish`, { method: 'POST' });
  }

  async attestPolicy(id: string) {
    return this.request<any>(`/policies/${id}/attest`, { method: 'POST' });
  }

  async getAttestationStats() {
    return this.request<any>('/policies/attestations/stats');
  }

  // ── Audits ────────────────────────────────────────
  async getAudits(page = 1, pageSize = 20) {
    return this.request<any>(`/audits?page=${page}&page_size=${pageSize}`);
  }

  async getAudit(id: string) {
    return this.request<any>(`/audits/${id}`);
  }

  async createAudit(data: any) {
    return this.request<any>('/audits', { method: 'POST', body: data });
  }

  async getAuditFindings(auditId: string) {
    return this.request<any>(`/audits/${auditId}/findings`);
  }

  async createFinding(auditId: string, data: any) {
    return this.request<any>(`/audits/${auditId}/findings`, { method: 'POST', body: data });
  }

  async getFindingsStats() {
    return this.request<any>('/audits/findings/stats');
  }

  // ── Incidents ─────────────────────────────────────
  async getIncidents(page = 1, pageSize = 20) {
    return this.request<any>(`/incidents?page=${page}&page_size=${pageSize}`);
  }

  async getIncident(id: string) {
    return this.request<any>(`/incidents/${id}`);
  }

  async reportIncident(data: any) {
    return this.request<any>('/incidents', { method: 'POST', body: data });
  }

  async notifyDPA(id: string) {
    return this.request<any>(`/incidents/${id}/notify-dpa`, { method: 'POST' });
  }

  async submitNIS2EarlyWarning(id: string) {
    return this.request<any>(`/incidents/${id}/nis2-early-warning`, { method: 'POST' });
  }

  async getIncidentStats() {
    return this.request<any>('/incidents/stats');
  }

  async getUrgentBreaches() {
    return this.request<any>('/incidents/breaches/urgent');
  }

  // ── Vendors ───────────────────────────────────────
  async getVendors(page = 1, pageSize = 20) {
    return this.request<any>(`/vendors?page=${page}&page_size=${pageSize}`);
  }

  async getVendor(id: string) {
    return this.request<any>(`/vendors/${id}`);
  }

  async onboardVendor(data: any) {
    return this.request<any>('/vendors', { method: 'POST', body: data });
  }

  async getVendorStats() {
    return this.request<any>('/vendors/stats');
  }

  // ── Assets ────────────────────────────────────────
  async getAssets(page = 1, pageSize = 20) {
    return this.request<any>(`/assets?page=${page}&page_size=${pageSize}`);
  }

  async registerAsset(data: any) {
    return this.request<any>('/assets', { method: 'POST', body: data });
  }

  async getAssetStats() {
    return this.request<any>('/assets/stats');
  }

  // ── Reports (Basic) ────────────────────────────────
  async getComplianceReport() {
    return this.request<any>('/reports/compliance');
  }

  async getRiskReport() {
    return this.request<any>('/reports/risk');
  }

  // ── Reports (Advanced — Prompt 12) ────────────────
  async generateReport(data: any) {
    return this.request<any>('/reports/generate', { method: 'POST', body: data });
  }

  async getReportStatus(id: string) {
    return this.request<any>(`/reports/status/${id}`);
  }

  async getReportDownloadUrl(id: string) {
    return `${API_BASE}/reports/download/${id}`;
  }

  async getReportDefinitions() {
    return this.request<any>('/reports/definitions');
  }

  async createReportDefinition(data: any) {
    return this.request<any>('/reports/definitions', { method: 'POST', body: data });
  }

  async updateReportDefinition(id: string, data: any) {
    return this.request<any>(`/reports/definitions/${id}`, { method: 'PUT', body: data });
  }

  async deleteReportDefinition(id: string) {
    return this.request<any>(`/reports/definitions/${id}`, { method: 'DELETE' });
  }

  async generateFromDefinition(id: string) {
    return this.request<any>(`/reports/definitions/${id}/generate`, { method: 'POST' });
  }

  async getReportSchedules() {
    return this.request<any>('/reports/schedules');
  }

  async createReportSchedule(data: any) {
    return this.request<any>('/reports/schedules', { method: 'POST', body: data });
  }

  async updateReportSchedule(id: string, data: any) {
    return this.request<any>(`/reports/schedules/${id}`, { method: 'PUT', body: data });
  }

  async deleteReportSchedule(id: string) {
    return this.request<any>(`/reports/schedules/${id}`, { method: 'DELETE' });
  }

  async getReportHistory(page = 1) {
    return this.request<any>(`/reports/history?page=${page}`);
  }

  // ── Notifications (Prompt 11) ─────────────────────
  async getNotifications(page = 1, pageSize = 10) {
    return this.request<any>(`/notifications?page=${page}&page_size=${pageSize}`);
  }

  async markNotificationRead(id: string) {
    return this.request<any>(`/notifications/${id}/read`, { method: 'PUT' });
  }

  async markAllNotificationsRead() {
    return this.request<any>('/notifications/read-all', { method: 'PUT' });
  }

  async getUnreadNotificationCount() {
    return this.request<any>('/notifications/unread-count');
  }

  async getNotificationPreferences() {
    return this.request<any>('/notifications/preferences');
  }

  async updateNotificationPreferences(data: any) {
    return this.request<any>('/notifications/preferences', { method: 'PUT', body: data });
  }

  async getNotificationRules() {
    return this.request<any>('/settings/notification-rules');
  }

  async createNotificationRule(data: any) {
    return this.request<any>('/settings/notification-rules', { method: 'POST', body: data });
  }

  async updateNotificationRule(id: string, data: any) {
    return this.request<any>(`/settings/notification-rules/${id}`, { method: 'PUT', body: data });
  }

  async deleteNotificationRule(id: string) {
    return this.request<any>(`/settings/notification-rules/${id}`, { method: 'DELETE' });
  }

  async getNotificationChannels() {
    return this.request<any>('/settings/notification-channels');
  }

  async createNotificationChannel(data: any) {
    return this.request<any>('/settings/notification-channels', { method: 'POST', body: data });
  }

  async testNotificationChannel(id: string) {
    return this.request<any>(`/settings/notification-channels/${id}/test`, { method: 'POST' });
  }

  // ── GDPR DSR (Prompt 13) ─────────────────────────
  async getDSRRequests(page = 1, pageSize = 20) {
    return this.request<any>(`/dsr?page=${page}&page_size=${pageSize}`);
  }

  async getDSRRequest(id: string) {
    return this.request<any>(`/dsr/${id}`);
  }

  async createDSRRequest(data: any) {
    return this.request<any>('/dsr', { method: 'POST', body: data });
  }

  async updateDSRRequest(id: string, data: any) {
    return this.request<any>(`/dsr/${id}`, { method: 'PUT', body: data });
  }

  async verifyDSRIdentity(id: string, data: any) {
    return this.request<any>(`/dsr/${id}/verify-identity`, { method: 'POST', body: data });
  }

  async assignDSR(id: string, data: any) {
    return this.request<any>(`/dsr/${id}/assign`, { method: 'POST', body: data });
  }

  async extendDSRDeadline(id: string, data: any) {
    return this.request<any>(`/dsr/${id}/extend`, { method: 'POST', body: data });
  }

  async completeDSR(id: string, data: any) {
    return this.request<any>(`/dsr/${id}/complete`, { method: 'POST', body: data });
  }

  async rejectDSR(id: string, data: any) {
    return this.request<any>(`/dsr/${id}/reject`, { method: 'POST', body: data });
  }

  async updateDSRTask(requestId: string, taskId: string, data: any) {
    return this.request<any>(`/dsr/${requestId}/tasks/${taskId}`, { method: 'PUT', body: data });
  }

  async getDSRDashboard() {
    return this.request<any>('/dsr/dashboard');
  }

  async getDSROverdue() {
    return this.request<any>('/dsr/overdue');
  }

  async getDSRTemplates() {
    return this.request<any>('/dsr/templates');
  }

  // ── NIS2 (Prompt 14) ─────────────────────────────
  async getNIS2Assessment() {
    return this.request<any>('/nis2/assessment');
  }

  async submitNIS2Assessment(data: any) {
    return this.request<any>('/nis2/assessment', { method: 'POST', body: data });
  }

  async getNIS2Incidents() {
    return this.request<any>('/nis2/incidents');
  }

  async getNIS2Incident(id: string) {
    return this.request<any>(`/nis2/incidents/${id}`);
  }

  async submitNIS2EarlyWarningReport(id: string, data: any) {
    return this.request<any>(`/nis2/incidents/${id}/early-warning`, { method: 'POST', body: data });
  }

  async submitNIS2Notification(id: string, data: any) {
    return this.request<any>(`/nis2/incidents/${id}/notification`, { method: 'POST', body: data });
  }

  async submitNIS2FinalReport(id: string, data: any) {
    return this.request<any>(`/nis2/incidents/${id}/final-report`, { method: 'POST', body: data });
  }

  async getNIS2Measures() {
    return this.request<any>('/nis2/measures');
  }

  async updateNIS2Measure(id: string, data: any) {
    return this.request<any>(`/nis2/measures/${id}`, { method: 'PUT', body: data });
  }

  async getNIS2Management() {
    return this.request<any>('/nis2/management');
  }

  async createNIS2ManagementRecord(data: any) {
    return this.request<any>('/nis2/management', { method: 'POST', body: data });
  }

  async getNIS2Dashboard() {
    return this.request<any>('/nis2/dashboard');
  }

  // ── Continuous Monitoring (Prompt 15) ─────────────
  async getMonitoringConfigs() {
    return this.request<any>('/monitoring/configs');
  }

  async createMonitoringConfig(data: any) {
    return this.request<any>('/monitoring/configs', { method: 'POST', body: data });
  }

  async updateMonitoringConfig(id: string, data: any) {
    return this.request<any>(`/monitoring/configs/${id}`, { method: 'PUT', body: data });
  }

  async runMonitoringConfigNow(id: string) {
    return this.request<any>(`/monitoring/configs/${id}/run-now`, { method: 'POST' });
  }

  async getMonitoringConfigHistory(id: string) {
    return this.request<any>(`/monitoring/configs/${id}/history`);
  }

  async getComplianceMonitors() {
    return this.request<any>('/monitoring/monitors');
  }

  async createComplianceMonitor(data: any) {
    return this.request<any>('/monitoring/monitors', { method: 'POST', body: data });
  }

  async updateComplianceMonitor(id: string, data: any) {
    return this.request<any>(`/monitoring/monitors/${id}`, { method: 'PUT', body: data });
  }

  async getMonitorResults(id: string) {
    return this.request<any>(`/monitoring/monitors/${id}/results`);
  }

  async getDriftEvents() {
    return this.request<any>('/monitoring/drift');
  }

  async acknowledgeDrift(id: string) {
    return this.request<any>(`/monitoring/drift/${id}/acknowledge`, { method: 'PUT' });
  }

  async resolveDrift(id: string, data: any) {
    return this.request<any>(`/monitoring/drift/${id}/resolve`, { method: 'PUT', body: data });
  }

  async getMonitoringDashboard() {
    return this.request<any>('/monitoring/dashboard');
  }

  // ── Settings ──────────────────────────────────────
  async getOrganization() {
    return this.request<any>('/settings');
  }

  async updateOrganization(data: any) {
    return this.request<any>('/settings', { method: 'PUT', body: data });
  }

  async getUsers(page = 1, search = '') {
    return this.request<any>(`/settings/users?page=${page}&search=${encodeURIComponent(search)}`);
  }

  async createUser(data: any) {
    return this.request<any>('/settings/users', { method: 'POST', body: data });
  }

  async getRoles() {
    return this.request<any>('/settings/roles');
  }

  async getAuditLog(page = 1) {
    return this.request<any>(`/settings/audit-log?page=${page}`);
  }

  // ── Workflows ─────────────────────────────────────
  async getWorkflowDefinitions(page = 1, pageSize = 50) {
    return this.request<any>(`/workflows/definitions?page=${page}&page_size=${pageSize}`);
  }

  async createWorkflowDefinition(data: any) {
    return this.request<any>('/workflows/definitions', { method: 'POST', body: data });
  }

  async updateWorkflowDefinition(id: string, data: any) {
    return this.request<any>(`/workflows/definitions/${id}`, { method: 'PUT', body: data });
  }

  async activateWorkflowDefinition(id: string) {
    return this.request<any>(`/workflows/definitions/${id}/activate`, { method: 'POST' });
  }

  async getWorkflowSteps(defId: string) {
    return this.request<any>(`/workflows/definitions/${defId}/steps`);
  }

  async addWorkflowStep(defId: string, data: any) {
    return this.request<any>(`/workflows/definitions/${defId}/steps`, { method: 'POST', body: data });
  }

  async updateWorkflowStep(defId: string, stepId: string, data: any) {
    return this.request<any>(`/workflows/definitions/${defId}/steps/${stepId}`, { method: 'PUT', body: data });
  }

  async deleteWorkflowStep(defId: string, stepId: string) {
    return this.request<any>(`/workflows/definitions/${defId}/steps/${stepId}`, { method: 'DELETE' });
  }

  async getWorkflowInstances(page = 1, filters: Record<string, string> = {}) {
    const params = new URLSearchParams({ page: String(page) });
    Object.entries(filters).forEach(([k, v]) => { if (v) params.set(k, v); });
    return this.request<any>(`/workflows/instances?${params.toString()}`);
  }

  async getWorkflowInstance(id: string) {
    return this.request<any>(`/workflows/instances/${id}`);
  }

  async startWorkflow(data: any) {
    return this.request<any>('/workflows/start', { method: 'POST', body: data });
  }

  async cancelWorkflow(id: string, reason: string) {
    return this.request<any>(`/workflows/instances/${id}/cancel`, { method: 'POST', body: { reason } });
  }

  async getMyApprovals() {
    return this.request<any>('/workflows/my-approvals');
  }

  async approveExecution(id: string, data: any = {}) {
    return this.request<any>(`/workflows/executions/${id}/approve`, { method: 'POST', body: data });
  }

  async rejectExecution(id: string, data: any) {
    return this.request<any>(`/workflows/executions/${id}/reject`, { method: 'POST', body: data });
  }

  async delegateExecution(id: string, data: any) {
    return this.request<any>(`/workflows/executions/${id}/delegate`, { method: 'POST', body: data });
  }

  async getWorkflowDelegations() {
    return this.request<any>('/workflows/delegations');
  }

  async createWorkflowDelegation(data: any) {
    return this.request<any>('/workflows/delegations', { method: 'POST', body: data });
  }

  // ── Access Control / ABAC (Prompt 20) ──────────────
  async getAccessPolicies(page = 1, pageSize = 50) {
    return this.request<any>(`/access/policies?page=${page}&page_size=${pageSize}`);
  }

  async createAccessPolicy(data: any) {
    return this.request<any>('/access/policies', { method: 'POST', body: data });
  }

  async updateAccessPolicy(id: string, data: any) {
    return this.request<any>(`/access/policies/${id}`, { method: 'PUT', body: data });
  }

  async deleteAccessPolicy(id: string) {
    return this.request<any>(`/access/policies/${id}`, { method: 'DELETE' });
  }

  async getAccessPolicyAssignments(policyId: string) {
    return this.request<any>(`/access/policies/${policyId}/assignments`);
  }

  async createAccessPolicyAssignment(policyId: string, data: any) {
    return this.request<any>(`/access/policies/${policyId}/assignments`, { method: 'POST', body: data });
  }

  async deleteAccessPolicyAssignment(policyId: string, assignmentId: string) {
    return this.request<any>(`/access/policies/${policyId}/assignments/${assignmentId}`, { method: 'DELETE' });
  }

  async evaluateAccessPolicy(data: any) {
    return this.request<any>('/access/evaluate', { method: 'POST', body: data });
  }

  async getAccessAuditLog(page = 1, pageSize = 50) {
    return this.request<any>(`/access/audit-log?page=${page}&page_size=${pageSize}`);
  }

  async getMyAccessPermissions() {
    return this.request<any>('/access/my-permissions');
  }

  async getFieldPermissions(resourceType: string) {
    return this.request<any>(`/access/field-permissions?resource_type=${resourceType}`);
  }

  // ── Integration Hub (Prompt 17) ──────────────────
  async getIntegrations() {
    return this.request<any>('/integrations');
  }

  async createIntegration(data: any) {
    return this.request<any>('/integrations', { method: 'POST', body: data });
  }

  async getIntegration(id: string) {
    return this.request<any>(`/integrations/${id}`);
  }

  async updateIntegration(id: string, data: any) {
    return this.request<any>(`/integrations/${id}`, { method: 'PUT', body: data });
  }

  async deleteIntegration(id: string) {
    return this.request<any>(`/integrations/${id}`, { method: 'DELETE' });
  }

  async testIntegration(id: string) {
    return this.request<any>(`/integrations/${id}/test`, { method: 'POST' });
  }

  async syncIntegration(id: string) {
    return this.request<any>(`/integrations/${id}/sync`, { method: 'POST' });
  }

  async getIntegrationSyncLogs(id: string, page = 1, pageSize = 20) {
    return this.request<any>(`/integrations/${id}/logs?page=${page}&page_size=${pageSize}`);
  }

  async getIntegrationHealth(id: string) {
    return this.request<any>(`/integrations/${id}/health`);
  }

  async getIntegrationHealthSummary() {
    return this.request<any>('/integrations/health/summary');
  }

  async getSSOConfig() {
    return this.request<any>('/settings/sso');
  }

  async updateSSOConfig(data: any) {
    return this.request<any>('/settings/sso', { method: 'PUT', body: data });
  }

  async getAPIKeys() {
    return this.request<any>('/settings/api-keys');
  }

  async createAPIKey(data: any) {
    return this.request<any>('/settings/api-keys', { method: 'POST', body: data });
  }

  async revokeAPIKey(id: string) {
    return this.request<any>(`/settings/api-keys/${id}`, { method: 'DELETE' });
  }

  // ── Regulatory Change Management (Prompt 23) ──────
  async getRegulatoryDashboard() {
    return this.request<any>('/regulatory/dashboard');
  }

  async getRegulatoryChanges(queryString = '') {
    return this.request<any>(`/regulatory/changes?${queryString}`);
  }

  async getRegulatoryChange(id: string) {
    return this.request<any>(`/regulatory/changes/${id}`);
  }

  async assessRegulatoryImpact(changeId: string, data: any) {
    return this.request<any>(`/regulatory/changes/${changeId}/assess`, { method: 'POST', body: data });
  }

  async getRegulatoryAssessment(changeId: string) {
    return this.request<any>(`/regulatory/changes/${changeId}/assessment`);
  }

  async createRegulatoryResponsePlan(changeId: string) {
    return this.request<any>(`/regulatory/changes/${changeId}/respond`, { method: 'POST' });
  }

  async getRegulatorySources() {
    return this.request<any>('/regulatory/sources');
  }

  async createRegulatorySource(data: any) {
    return this.request<any>('/regulatory/sources', { method: 'POST', body: data });
  }

  async getRegulatorySubscriptions() {
    return this.request<any>('/regulatory/subscriptions');
  }

  async subscribeRegulatory(data: any) {
    return this.request<any>('/regulatory/subscriptions', { method: 'POST', body: data });
  }

  async unsubscribeRegulatory(id: string) {
    return this.request<any>(`/regulatory/subscriptions/${id}`, { method: 'DELETE' });
  }

  async getRegulatoryTimeline(months = 6) {
    return this.request<any>(`/regulatory/timeline?months=${months}`);
  }

  // ── Marketplace (Prompt 22) ────────────────────────
  async getMarketplacePackages(queryString = '') {
    return this.request<any>(`/marketplace/packages?${queryString}`);
  }

  async getMarketplaceFeatured() {
    return this.request<any>('/marketplace/packages/featured');
  }

  async getMarketplacePackagesByFramework(code: string) {
    return this.request<any>(`/marketplace/packages/framework/${code}`);
  }

  async getMarketplacePackageDetail(publisher: string, slug: string) {
    return this.request<any>(`/marketplace/packages/${publisher}/${slug}`);
  }

  async getMarketplacePackageVersions(publisher: string, slug: string) {
    return this.request<any>(`/marketplace/packages/${publisher}/${slug}/versions`);
  }

  async getMarketplacePackageReviews(publisher: string, slug: string, page = 1, pageSize = 10) {
    return this.request<any>(`/marketplace/packages/${publisher}/${slug}/reviews?page=${page}&page_size=${pageSize}`);
  }

  async installMarketplacePackage(data: any) {
    return this.request<any>('/marketplace/install', { method: 'POST', body: data });
  }

  async uninstallMarketplacePackage(packageId: string) {
    return this.request<any>(`/marketplace/install/${packageId}`, { method: 'DELETE' });
  }

  async updateMarketplaceInstallation(installationId: string) {
    return this.request<any>(`/marketplace/install/${installationId}/update`, { method: 'POST' });
  }

  async getMarketplaceInstalled() {
    return this.request<any>('/marketplace/installed');
  }

  async submitMarketplaceReview(data: any) {
    return this.request<any>('/marketplace/reviews', { method: 'POST', body: data });
  }

  async createMarketplacePublisher(data: any) {
    return this.request<any>('/marketplace/publishers', { method: 'POST', body: data });
  }

  async getMyPublisherProfile() {
    return this.request<any>('/marketplace/publishers/me');
  }

  async getMyPublisherStats() {
    return this.request<any>('/marketplace/publishers/me/stats');
  }

  async createMarketplacePackage(data: any) {
    return this.request<any>('/marketplace/publishers/me/packages', { method: 'POST', body: data });
  }

  async publishMarketplaceVersion(packageId: string, data: any) {
    return this.request<any>(`/marketplace/publishers/me/packages/${packageId}/versions`, { method: 'POST', body: data });
  }

  async exportAsMarketplacePackage(data: any) {
    return this.request<any>('/marketplace/export', { method: 'POST', body: data });
  }

  // ── Exception Management (Prompt 26) ───────────
  async getExceptions(queryString = '') {
    return this.request<any>(`/exceptions?${queryString}`);
  }

  async getException(id: string) {
    return this.request<any>(`/exceptions/${id}`);
  }

  async createException(data: any) {
    return this.request<any>('/exceptions', { method: 'POST', body: data });
  }

  async updateException(id: string, data: any) {
    return this.request<any>(`/exceptions/${id}`, { method: 'PUT', body: data });
  }

  async submitExceptionForApproval(id: string) {
    return this.request<any>(`/exceptions/${id}/submit`, { method: 'POST' });
  }

  async approveException(id: string, data: any) {
    return this.request<any>(`/exceptions/${id}/approve`, { method: 'POST', body: data });
  }

  async rejectException(id: string, data: any) {
    return this.request<any>(`/exceptions/${id}/reject`, { method: 'POST', body: data });
  }

  async revokeException(id: string, data: any) {
    return this.request<any>(`/exceptions/${id}/revoke`, { method: 'POST', body: data });
  }

  async renewException(id: string, data: any) {
    return this.request<any>(`/exceptions/${id}/renew`, { method: 'POST', body: data });
  }

  async reviewException(id: string, data: any) {
    return this.request<any>(`/exceptions/${id}/review`, { method: 'POST', body: data });
  }

  async getExceptionReviews(id: string) {
    return this.request<any>(`/exceptions/${id}/reviews`);
  }

  async getExceptionAuditTrail(id: string) {
    return this.request<any>(`/exceptions/${id}/audit-trail`);
  }

  async getExceptionDashboard() {
    return this.request<any>('/exceptions/dashboard');
  }

  async getExpiringExceptions(days = 30) {
    return this.request<any>(`/exceptions/expiring?days=${days}`);
  }

  async getExceptionComplianceImpact(id: string) {
    return this.request<any>(`/exceptions/impact/${id}`);
  }

  // ── Board Reporting & Governance (Prompt 30) ──
  async getBoardDashboard() {
    return this.request<any>('/board/dashboard');
  }

  async getBoardMembers() {
    return this.request<any>('/board/members');
  }

  async createBoardMember(data: any) {
    return this.request<any>('/board/members', { method: 'POST', body: data });
  }

  async updateBoardMember(id: string, data: any) {
    return this.request<any>(`/board/members/${id}`, { method: 'PUT', body: data });
  }

  async getBoardMeetings(queryString = '') {
    return this.request<any>(`/board/meetings?${queryString}`);
  }

  async createBoardMeeting(data: any) {
    return this.request<any>('/board/meetings', { method: 'POST', body: data });
  }

  async getBoardMeeting(id: string) {
    return this.request<any>(`/board/meetings/${id}`);
  }

  async updateBoardMeeting(id: string, data: any) {
    return this.request<any>(`/board/meetings/${id}`, { method: 'PUT', body: data });
  }

  async generateBoardPack(meetingId: string) {
    return this.request<any>(`/board/meetings/${meetingId}/generate-pack`, { method: 'POST' });
  }

  async downloadBoardPack(meetingId: string) {
    return this.request<any>(`/board/meetings/${meetingId}/download-pack`);
  }

  async recordBoardDecision(data: any) {
    return this.request<any>('/board/decisions', { method: 'POST', body: data });
  }

  async getBoardDecisions(queryString = '') {
    return this.request<any>(`/board/decisions?${queryString}`);
  }

  async updateBoardDecisionAction(decisionId: string, data: any) {
    return this.request<any>(`/board/decisions/${decisionId}/action`, { method: 'PUT', body: data });
  }

  async getBoardReports() {
    return this.request<any>('/board/reports');
  }

  async generateBoardReport(data: any) {
    return this.request<any>('/board/reports/generate', { method: 'POST', body: data });
  }

  async getNIS2GovernanceReport() {
    return this.request<any>('/board/nis2-governance');
  }

  // ── TPRM — Questionnaires (Prompt 28) ──────────
  async getQuestionnaires(page = 1, pageSize = 20, queryString = '') {
    return this.request<any>(`/questionnaires?page=${page}&page_size=${pageSize}&${queryString}`);
  }

  async getQuestionnaire(id: string) {
    return this.request<any>(`/questionnaires/${id}`);
  }

  async createQuestionnaire(data: any) {
    return this.request<any>('/questionnaires', { method: 'POST', body: data });
  }

  async updateQuestionnaire(id: string, data: any) {
    return this.request<any>(`/questionnaires/${id}`, { method: 'PUT', body: data });
  }

  async cloneQuestionnaire(id: string, data: any) {
    return this.request<any>(`/questionnaires/${id}/clone`, { method: 'POST', body: data });
  }

  // ── TPRM — Vendor Assessments (Prompt 28) ──────
  async getVendorAssessments(page = 1, pageSize = 20, queryString = '') {
    return this.request<any>(`/vendor-assessments?page=${page}&page_size=${pageSize}&${queryString}`);
  }

  async getVendorAssessment(id: string) {
    return this.request<any>(`/vendor-assessments/${id}`);
  }

  async sendVendorAssessment(data: any) {
    return this.request<any>('/vendor-assessments', { method: 'POST', body: data });
  }

  async getVendorAssessmentDashboard() {
    return this.request<any>('/vendor-assessments/dashboard');
  }

  async compareVendorAssessments(ids: string[]) {
    return this.request<any>(`/vendor-assessments/compare?ids=${ids.join(',')}`);
  }

  async reviewVendorAssessment(id: string, data: any) {
    return this.request<any>(`/vendor-assessments/${id}/review`, { method: 'POST', body: data });
  }

  async sendVendorAssessmentReminder(id: string) {
    return this.request<any>(`/vendor-assessments/${id}/reminder`, { method: 'POST' });
  }

  // ── Evidence Template Library & Testing (Prompt 27) ──
  async getEvidenceTemplates(queryString = '') {
    return this.request<any>(`/evidence/templates?${queryString}`);
  }

  async getEvidenceTemplate(id: string) {
    return this.request<any>(`/evidence/templates/${id}`);
  }

  async createEvidenceTemplate(data: any) {
    return this.request<any>('/evidence/templates', { method: 'POST', body: data });
  }

  async getEvidenceRequirements(queryString = '') {
    return this.request<any>(`/evidence/requirements?${queryString}`);
  }

  async generateEvidenceRequirements(data: any) {
    return this.request<any>('/evidence/requirements/generate', { method: 'POST', body: data });
  }

  async updateEvidenceRequirement(id: string, data: any) {
    return this.request<any>(`/evidence/requirements/${id}`, { method: 'PUT', body: data });
  }

  async validateEvidence(requirementId: string, data: any) {
    return this.request<any>(`/evidence/requirements/${requirementId}/validate`, { method: 'POST', body: data });
  }

  async getEvidenceGaps() {
    return this.request<any>('/evidence/gaps');
  }

  async getCollectionSchedule() {
    return this.request<any>('/evidence/schedule');
  }

  async getEvidenceTestSuites() {
    return this.request<any>('/evidence/test-suites');
  }

  async createEvidenceTestSuite(data: any) {
    return this.request<any>('/evidence/test-suites', { method: 'POST', body: data });
  }

  async runEvidenceTestSuite(suiteId: string) {
    return this.request<any>(`/evidence/test-suites/${suiteId}/run`, { method: 'POST' });
  }

  async getEvidenceTestRunResults(suiteId: string) {
    return this.request<any>(`/evidence/test-suites/${suiteId}/results`);
  }

  async runPreAuditCheck(data: any) {
    return this.request<any>('/evidence/pre-audit-check', { method: 'POST', body: data });
  }

  async getPreAuditReport(id: string) {
    return this.request<any>(`/evidence/pre-audit-check/${id}/report`);
  }

  // ── Search & Knowledge Base (Prompt 32) ─────────
  async searchEntities(queryString: string) {
    return this.request<any>(`/search?${queryString}`);
  }

  async searchAutocomplete(query: string, limit = 10) {
    return this.request<any>(`/search/autocomplete?q=${encodeURIComponent(query)}&limit=${limit}`);
  }

  async getRelatedEntities(entityType: string, entityId: string) {
    return this.request<any>(`/search/related/${entityType}/${entityId}`);
  }

  async getRecentSearches(limit = 10) {
    return this.request<any>(`/search/recent?limit=${limit}`);
  }

  async getSearchAnalytics(days = 30) {
    return this.request<any>(`/search/analytics?days=${days}`);
  }

  async getIndexStats() {
    return this.request<any>('/search/index-stats');
  }

  async triggerReindex() {
    return this.request<any>('/search/reindex', { method: 'POST' });
  }

  async recordSearchClick(data: { query: string; entity_type: string; entity_id: string }) {
    return this.request<any>('/search/click', { method: 'POST', body: data });
  }

  async browseKnowledge(queryString: string) {
    return this.request<any>(`/knowledge?${queryString}`);
  }

  async getKnowledgeArticle(slug: string) {
    return this.request<any>(`/knowledge/${slug}`);
  }

  async getArticlesForControl(frameworkCode: string, controlCode: string) {
    return this.request<any>(`/knowledge/for-control/${frameworkCode}/${controlCode}`);
  }

  async getRecommendedArticles(limit = 5) {
    return this.request<any>(`/knowledge/recommended?limit=${limit}`);
  }

  async createKnowledgeArticle(data: any) {
    return this.request<any>('/knowledge/articles', { method: 'POST', body: data });
  }

  async updateKnowledgeArticle(id: string, data: any) {
    return this.request<any>(`/knowledge/articles/${id}`, { method: 'PUT', body: data });
  }

  async submitArticleFeedback(id: string, data: { action: string; comment?: string }) {
    return this.request<any>(`/knowledge/articles/${id}/feedback`, { method: 'POST', body: data });
  }

  async getKnowledgeBookmarks(page = 1, pageSize = 20) {
    return this.request<any>(`/knowledge/bookmarks?page=${page}&page_size=${pageSize}`);
  }

  async toggleKnowledgeBookmark(articleId: string) {
    return this.request<any>(`/knowledge/bookmarks/${articleId}`, { method: 'POST' });
  }

  // ── Collaboration — Comments (Prompt 33) ────────
  async getComments(entityType: string, entityId: string, sort = 'oldest') {
    return this.request<any>(`/comments/${entityType}/${entityId}?sort=${sort}`);
  }

  async createComment(entityType: string, entityId: string, data: any) {
    return this.request<any>(`/comments/${entityType}/${entityId}`, { method: 'POST', body: data });
  }

  async editComment(commentId: string, data: any) {
    return this.request<any>(`/comments/${commentId}`, { method: 'PUT', body: data });
  }

  async deleteComment(commentId: string) {
    return this.request<any>(`/comments/${commentId}`, { method: 'DELETE' });
  }

  async pinComment(commentId: string) {
    return this.request<any>(`/comments/${commentId}/pin`, { method: 'POST' });
  }

  async reactToComment(commentId: string, reactionType: string) {
    return this.request<any>(`/comments/${commentId}/react`, { method: 'POST', body: { reaction_type: reactionType } });
  }

  // ── Collaboration — Activity Feed (Prompt 33) ──
  async getActivityFeed(page = 1, pageSize = 20, filters: Record<string, string> = {}) {
    const params = new URLSearchParams({ page: String(page), page_size: String(pageSize) });
    Object.entries(filters).forEach(([k, v]) => { if (v) params.set(k, v); });
    return this.request<any>(`/activity/feed?${params.toString()}`);
  }

  async getOrgActivityFeed(page = 1, pageSize = 20, filters: Record<string, string> = {}) {
    const params = new URLSearchParams({ page: String(page), page_size: String(pageSize) });
    Object.entries(filters).forEach(([k, v]) => { if (v) params.set(k, v); });
    return this.request<any>(`/activity/org?${params.toString()}`);
  }

  async getEntityActivity(entityType: string, entityId: string, page = 1, pageSize = 20) {
    return this.request<any>(`/activity/${entityType}/${entityId}?page=${page}&page_size=${pageSize}`);
  }

  async getUnreadCounts() {
    return this.request<any>('/activity/unread');
  }

  async markEntityRead(entityType: string, entityId: string) {
    return this.request<any>(`/activity/${entityType}/${entityId}/mark-read`, { method: 'POST' });
  }

  // ── Collaboration — Following (Prompt 33) ──────
  async getFollowing(page = 1, pageSize = 20) {
    return this.request<any>(`/following?page=${page}&page_size=${pageSize}`);
  }

  async followEntity(entityType: string, entityId: string, followType = 'watching') {
    return this.request<any>(`/following/${entityType}/${entityId}`, { method: 'POST', body: { follow_type: followType } });
  }

  async unfollowEntity(entityType: string, entityId: string) {
    return this.request<any>(`/following/${entityType}/${entityId}`, { method: 'DELETE' });
  }

  // ── Mobile API (Prompt 34) ────────────────────────
  async getMobileDashboard() {
    return this.request<any>('/mobile/dashboard');
  }

  async getMobileApprovals(page = 1, pageSize = 20) {
    return this.request<any>(`/mobile/approvals?page=${page}&page_size=${pageSize}`);
  }

  async mobileApprove(id: string, comment = '') {
    return this.request<any>(`/mobile/approvals/${id}/approve`, { method: 'POST', body: { comment } });
  }

  async mobileReject(id: string, reason: string) {
    return this.request<any>(`/mobile/approvals/${id}/reject`, { method: 'POST', body: { reason } });
  }

  async getMobileIncidents(page = 1, pageSize = 20) {
    return this.request<any>(`/mobile/incidents/active?page=${page}&page_size=${pageSize}`);
  }

  async getMobileDeadlines(days = 7) {
    return this.request<any>(`/mobile/deadlines?days=${days}`);
  }

  async getMobileActivity(limit = 20) {
    return this.request<any>(`/mobile/activity?limit=${limit}`);
  }

  async registerPushToken(data: { platform: string; token: string; device_name?: string; device_model?: string; os_version?: string; app_version?: string }) {
    return this.request<any>('/mobile/push/register', { method: 'POST', body: data });
  }

  async unregisterPushToken(data: { token_hash?: string; token?: string }) {
    return this.request<any>('/mobile/push/unregister', { method: 'DELETE', body: data });
  }

  async getPushPreferences() {
    return this.request<any>('/mobile/push/preferences');
  }

  async updatePushPreferences(data: any) {
    return this.request<any>('/mobile/push/preferences', { method: 'PUT', body: data });
  }

  // ── Branding & White-Labelling (Prompt 35) ──────
  async getBranding() {
    return this.request<any>('/branding');
  }

  async getBrandingCSS() {
    const token = this.getToken();
    const headers: Record<string, string> = {};
    if (token) headers['Authorization'] = `Bearer ${token}`;
    const response = await fetch(`${API_BASE}/branding/css`, { headers });
    return response.text();
  }

  async updateBranding(data: any) {
    return this.request<any>('/branding', { method: 'PUT', body: data });
  }

  async uploadLogo(logoType: string, file: File) {
    const formData = new FormData();
    formData.append('logo_type', logoType);
    formData.append('file', file);

    const token = this.getToken();
    const headers: Record<string, string> = {};
    if (token) headers['Authorization'] = `Bearer ${token}`;

    const response = await fetch(`${API_BASE}/branding/logo`, {
      method: 'POST',
      headers,
      body: formData,
    });

    if (!response.ok) {
      const data = await response.json();
      throw new Error(data.error?.message || 'Upload failed');
    }
    return response.json();
  }

  async deleteLogo(logoType: string) {
    return this.request<any>(`/branding/logo/${logoType}`, { method: 'DELETE' });
  }

  async verifyDomain(data: { domain: string }) {
    return this.request<any>('/branding/domain/verify', { method: 'POST', body: data });
  }

  async getDomainStatus() {
    return this.request<any>('/branding/domain/status');
  }

  async previewBranding(data: any) {
    return this.request<any>('/branding/preview', { method: 'POST', body: data });
  }

  // ── White-Label Partners (Super Admin, Prompt 35) ──
  async getPartners() {
    return this.request<any>('/admin/partners');
  }

  async createPartner(data: any) {
    return this.request<any>('/admin/partners', { method: 'POST', body: data });
  }

  async updatePartner(id: string, data: any) {
    return this.request<any>(`/admin/partners/${id}`, { method: 'PUT', body: data });
  }

  async getPartnerTenants(id: string) {
    return this.request<any>(`/admin/partners/${id}/tenants`);
  }

  // ══════════════════════════════════════════════════
  // BATCH 8 — Prompts 36–40
  // ══════════════════════════════════════════════════

  // ── Compliance-as-Code Engine (Prompt 36) ─────────
  async listCaCRepositories(page = 1, pageSize = 20) {
    return this.request<any>(`/cac/repositories?page=${page}&page_size=${pageSize}`);
  }

  async createCaCRepository(data: any) {
    return this.request<any>('/cac/repositories', { method: 'POST', body: data });
  }

  async updateCaCRepository(id: string, data: any) {
    return this.request<any>(`/cac/repositories/${id}`, { method: 'PUT', body: data });
  }

  async deleteCaCRepository(id: string) {
    return this.request<any>(`/cac/repositories/${id}`, { method: 'DELETE' });
  }

  async triggerCaCSync(repoId: string) {
    return this.request<any>(`/cac/repositories/${repoId}/sync`, { method: 'POST' });
  }

  async getCaCRepoStatus(repoId: string) {
    return this.request<any>(`/cac/repositories/${repoId}/status`);
  }

  async listCaCSyncRuns(page = 1, pageSize = 20) {
    return this.request<any>(`/cac/sync-runs?page=${page}&page_size=${pageSize}`);
  }

  async getCaCSyncRun(id: string) {
    return this.request<any>(`/cac/sync-runs/${id}`);
  }

  async approveCaCSyncRun(id: string) {
    return this.request<any>(`/cac/sync-runs/${id}/approve`, { method: 'POST' });
  }

  async rejectCaCSyncRun(id: string, reason: string) {
    return this.request<any>(`/cac/sync-runs/${id}/reject`, { method: 'POST', body: { reason } });
  }

  async listCaCDriftEvents(page = 1, pageSize = 20, filters?: { direction?: string; status?: string; kind?: string }) {
    const params = new URLSearchParams({ page: String(page), page_size: String(pageSize) });
    if (filters?.direction) params.set('direction', filters.direction);
    if (filters?.status) params.set('status', filters.status);
    if (filters?.kind) params.set('kind', filters.kind);
    return this.request<any>(`/cac/drift?${params}`);
  }

  async resolveCaCDrift(id: string, resolution: string) {
    return this.request<any>(`/cac/drift/${id}/resolve`, { method: 'POST', body: { resolution } });
  }

  async listCaCResourceMappings(page = 1, pageSize = 20, filters?: { kind?: string; status?: string; search?: string }) {
    const params = new URLSearchParams({ page: String(page), page_size: String(pageSize) });
    if (filters?.kind) params.set('kind', filters.kind);
    if (filters?.status) params.set('status', filters.status);
    if (filters?.search) params.set('search', filters.search);
    return this.request<any>(`/cac/resource-mappings?${params}`);
  }

  async validateCaCYAML(content: string) {
    return this.request<any>('/cac/validate', { method: 'POST', body: { content } });
  }

  async planCaCChanges(repoId: string, content: string) {
    return this.request<any>('/cac/plan', { method: 'POST', body: { repository_id: repoId, content } });
  }

  async applyCaCChanges(syncRunId: string) {
    return this.request<any>('/cac/apply', { method: 'POST', body: { sync_run_id: syncRunId } });
  }

  async exportCaCYAML() {
    return this.request<any>('/cac/export', { method: 'POST' });
  }

  // ── Data Residency (Prompt 37) ────────────────────
  async getResidencyConfig() {
    return this.request<any>('/residency/config');
  }

  async getResidencyStatus() {
    return this.request<any>('/residency/status');
  }

  async getResidencyAuditLog(page = 1, pageSize = 20, filters?: { action?: string; allowed?: string; date_from?: string; date_to?: string }) {
    const params = new URLSearchParams({ page: String(page), page_size: String(pageSize) });
    if (filters?.action) params.set('action', filters.action);
    if (filters?.allowed) params.set('allowed', filters.allowed);
    if (filters?.date_from) params.set('date_from', filters.date_from);
    if (filters?.date_to) params.set('date_to', filters.date_to);
    return this.request<any>(`/residency/audit-log?${params}`);
  }

  async validateDataExport(destinationRegion: string) {
    return this.request<any>('/residency/validate-export', { method: 'POST', body: { destination_region: destinationRegion } });
  }

  async validateDataTransfer(vendorId: string, destinationCountry: string) {
    return this.request<any>('/residency/validate-transfer', { method: 'POST', body: { vendor_id: vendorId, destination_country: destinationCountry } });
  }

  async listResidencyRegions() {
    return this.request<any>('/residency/regions');
  }

  // ── Advanced Audit Management (Prompt 38) ─────────
  async listAuditProgrammes(page = 1, pageSize = 20) {
    return this.request<any>(`/audit-programmes/programmes?page=${page}&page_size=${pageSize}`);
  }

  async createAuditProgramme(data: any) {
    return this.request<any>('/audit-programmes/programmes', { method: 'POST', body: data });
  }

  async getAuditProgramme(id: string) {
    return this.request<any>(`/audit-programmes/programmes/${id}`);
  }

  async triggerRiskBasedSelection(programmeId: string, data: any) {
    return this.request<any>(`/audit-programmes/programmes/${programmeId}/risk-selection`, { method: 'POST', body: data });
  }

  async listAuditUniverse(page = 1, pageSize = 20) {
    return this.request<any>(`/audit-programmes/universe?page=${page}&page_size=${pageSize}`);
  }

  async createAuditableEntity(data: any) {
    return this.request<any>('/audit-programmes/universe', { method: 'POST', body: data });
  }

  async updateAuditableEntity(id: string, data: any) {
    return this.request<any>(`/audit-programmes/universe/${id}`, { method: 'PUT', body: data });
  }

  async listAuditEngagements(page = 1, pageSize = 20) {
    return this.request<any>(`/audit-programmes/engagements?page=${page}&page_size=${pageSize}`);
  }

  async createAuditEngagement(data: any) {
    return this.request<any>('/audit-programmes/engagements', { method: 'POST', body: data });
  }

  async getAuditEngagement(id: string) {
    return this.request<any>(`/audit-programmes/engagements/${id}`);
  }

  async updateEngagementStatus(id: string, status: string) {
    return this.request<any>(`/audit-programmes/engagements/${id}/status`, { method: 'PUT', body: { status } });
  }

  async listWorkpapers(engagementId: string) {
    return this.request<any>(`/audit-programmes/engagements/${engagementId}/workpapers`);
  }

  async createWorkpaper(engagementId: string, data: any) {
    return this.request<any>(`/audit-programmes/engagements/${engagementId}/workpapers`, { method: 'POST', body: data });
  }

  async updateWorkpaper(id: string, data: any) {
    return this.request<any>(`/audit-programmes/workpapers/${id}`, { method: 'PUT', body: data });
  }

  async submitWorkpaperForReview(id: string) {
    return this.request<any>(`/audit-programmes/workpapers/${id}/review`, { method: 'POST' });
  }

  async generateAuditSample(engagementId: string, config: any) {
    return this.request<any>(`/audit-programmes/engagements/${engagementId}/samples`, { method: 'POST', body: config });
  }

  async getAuditSample(id: string) {
    return this.request<any>(`/audit-programmes/samples/${id}`);
  }

  async recordSampleItemResult(sampleId: string, itemIndex: number, data: any) {
    return this.request<any>(`/audit-programmes/samples/${sampleId}/items/${itemIndex}`, { method: 'PUT', body: data });
  }

  async createTestProcedure(engagementId: string, data: any) {
    return this.request<any>(`/audit-programmes/engagements/${engagementId}/test-procedures`, { method: 'POST', body: data });
  }

  async recordTestResult(testProcedureId: string, data: any) {
    return this.request<any>(`/audit-programmes/test-procedures/${testProcedureId}`, { method: 'PUT', body: data });
  }

  async listCorrectiveActions(page = 1, pageSize = 20, filters?: { status?: string; priority?: string }) {
    const params = new URLSearchParams({ page: String(page), page_size: String(pageSize) });
    if (filters?.status) params.set('status', filters.status);
    if (filters?.priority) params.set('priority', filters.priority);
    return this.request<any>(`/audit-programmes/corrective-actions?${params}`);
  }

  async createCorrectiveAction(data: any) {
    return this.request<any>('/audit-programmes/corrective-actions', { method: 'POST', body: data });
  }

  async updateCorrectiveAction(id: string, data: any) {
    return this.request<any>(`/audit-programmes/corrective-actions/${id}`, { method: 'PUT', body: data });
  }

  async verifyCorrectiveAction(id: string) {
    return this.request<any>(`/audit-programmes/corrective-actions/${id}/verify`, { method: 'POST' });
  }

  // ── Training & Certification (Prompt 39) ──────────
  async listTrainingProgrammes(page = 1, pageSize = 20) {
    return this.request<any>(`/training/programmes?page=${page}&page_size=${pageSize}`);
  }

  async createTrainingProgramme(data: any) {
    return this.request<any>('/training/programmes', { method: 'POST', body: data });
  }

  async getTrainingProgramme(id: string) {
    return this.request<any>(`/training/programmes/${id}`);
  }

  async updateTrainingProgramme(id: string, data: any) {
    return this.request<any>(`/training/programmes/${id}`, { method: 'PUT', body: data });
  }

  async generateTrainingAssignments(programmeId: string) {
    return this.request<any>(`/training/programmes/${programmeId}/generate-assignments`, { method: 'POST' });
  }

  async getMyTrainingAssignments() {
    return this.request<any>('/training/my-assignments');
  }

  async listTrainingAssignments(page = 1, pageSize = 20) {
    return this.request<any>(`/training/assignments?page=${page}&page_size=${pageSize}`);
  }

  async startTrainingAssignment(id: string) {
    return this.request<any>(`/training/assignments/${id}/start`, { method: 'POST' });
  }

  async completeTrainingAssignment(id: string, data: { score: number; time_spent_minutes: number }) {
    return this.request<any>(`/training/assignments/${id}/complete`, { method: 'POST', body: data });
  }

  async exemptTrainingAssignment(id: string, reason: string) {
    return this.request<any>(`/training/assignments/${id}/exempt`, { method: 'POST', body: { reason } });
  }

  async getTrainingCertificate(assignmentId: string) {
    return this.request<any>(`/training/assignments/${assignmentId}/certificate`);
  }

  async getTrainingDashboard() {
    return this.request<any>('/training/dashboard');
  }

  async getTrainingComplianceMatrix(department?: string) {
    const params = department ? `?department=${department}` : '';
    return this.request<any>(`/training/compliance-matrix${params}`);
  }

  async exportTrainingComplianceMatrix(department?: string) {
    const params = department ? `?department=${department}` : '';
    const token = this.getToken();
    const headers: Record<string, string> = {};
    if (token) headers['Authorization'] = `Bearer ${token}`;
    const response = await fetch(`${API_BASE}/training/compliance-matrix/export${params}`, { headers });
    return response.text();
  }

  async listPhishingSimulations(page = 1, pageSize = 20) {
    return this.request<any>(`/training/phishing/simulations?page=${page}&page_size=${pageSize}`);
  }

  async createPhishingSimulation(data: any) {
    return this.request<any>('/training/phishing/simulations', { method: 'POST', body: data });
  }

  async launchPhishingSimulation(id: string) {
    return this.request<any>(`/training/phishing/simulations/${id}/launch`, { method: 'POST' });
  }

  async getPhishingSimulationResults(id: string) {
    return this.request<any>(`/training/phishing/simulations/${id}/results`);
  }

  async getPhishingTrend() {
    return this.request<any>('/training/phishing/trend');
  }

  async listProfessionalCertifications(page = 1, pageSize = 20) {
    return this.request<any>(`/training/certifications?page=${page}&page_size=${pageSize}`);
  }

  async addProfessionalCertification(data: any) {
    return this.request<any>('/training/certifications', { method: 'POST', body: data });
  }

  async updateProfessionalCertification(id: string, data: any) {
    return this.request<any>(`/training/certifications/${id}`, { method: 'PUT', body: data });
  }

  async getExpiringCertifications(withinDays = 90) {
    return this.request<any>(`/training/certifications/expiring?within_days=${withinDays}`);
  }

  async getCertificationMatrix() {
    return this.request<any>('/training/certifications/matrix');
  }

  // ── Developer Portal & Webhooks (Prompt 40) ──────
  async listAPIKeys(page = 1, pageSize = 20) {
    return this.request<any>(`/developer/api-keys?page=${page}&page_size=${pageSize}`);
  }

  async generateAPIKey(data: any) {
    return this.request<any>('/developer/api-keys', { method: 'POST', body: data });
  }

  async updateAPIKey(id: string, data: any) {
    return this.request<any>(`/developer/api-keys/${id}`, { method: 'PUT', body: data });
  }

  async revokeAPIKey(id: string) {
    return this.request<any>(`/developer/api-keys/${id}`, { method: 'DELETE' });
  }

  async getAPIKeyUsage(id: string, period = '7d') {
    return this.request<any>(`/developer/api-keys/${id}/usage?period=${period}`);
  }

  async listWebhookSubscriptions(page = 1, pageSize = 20) {
    return this.request<any>(`/developer/webhooks?page=${page}&page_size=${pageSize}`);
  }

  async createWebhookSubscription(data: any) {
    return this.request<any>('/developer/webhooks', { method: 'POST', body: data });
  }

  async updateWebhookSubscription(id: string, data: any) {
    return this.request<any>(`/developer/webhooks/${id}`, { method: 'PUT', body: data });
  }

  async deleteWebhookSubscription(id: string) {
    return this.request<any>(`/developer/webhooks/${id}`, { method: 'DELETE' });
  }

  async testWebhook(id: string) {
    return this.request<any>(`/developer/webhooks/${id}/test`, { method: 'POST' });
  }

  async getWebhookDeliveries(subscriptionId: string, page = 1, pageSize = 20) {
    return this.request<any>(`/developer/webhooks/${subscriptionId}/deliveries?page=${page}&page_size=${pageSize}`);
  }

  async replayWebhookDelivery(deliveryId: string) {
    return this.request<any>(`/developer/webhooks/deliveries/${deliveryId}/replay`, { method: 'POST' });
  }

  async createSandbox() {
    return this.request<any>('/developer/sandbox', { method: 'POST' });
  }

  async getSandbox() {
    return this.request<any>('/developer/sandbox');
  }

  async destroySandbox() {
    return this.request<any>('/developer/sandbox', { method: 'DELETE' });
  }

  async listWebhookEventTypes() {
    return this.request<any>('/developer/events');
  }

  async listAPIScopes() {
    return this.request<any>('/developer/scopes');
  }
}

export const api = new ApiClient();
export default api;
