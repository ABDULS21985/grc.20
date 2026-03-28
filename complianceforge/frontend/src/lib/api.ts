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
}

export const api = new ApiClient();
export default api;
