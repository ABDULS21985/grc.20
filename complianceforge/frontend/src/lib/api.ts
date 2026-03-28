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

  // ── Reports ───────────────────────────────────────
  async getComplianceReport() {
    return this.request<any>('/reports/compliance');
  }

  async getRiskReport() {
    return this.request<any>('/reports/risk');
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
}

export const api = new ApiClient();
export default api;
