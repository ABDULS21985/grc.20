'use client';

const FRAMEWORKS = [
  { code: 'ISO27001', name: 'ISO/IEC 27001:2022', version: '2022', body: 'ISO/IEC', controls: 93, color: '#1E88E5', category: 'Security', description: 'Information Security Management Systems — the gold standard for ISMS.' },
  { code: 'UK_GDPR', name: 'UK GDPR', version: '2021', body: 'UK Government / ICO', controls: 36, color: '#7B1FA2', category: 'Privacy', description: 'UK data protection regulation governing personal data processing.' },
  { code: 'NCSC_CAF', name: 'NCSC CAF', version: '3.2', body: 'NCSC (UK)', controls: 14, color: '#FF6F00', category: 'Security', description: 'Cyber Assessment Framework for critical national infrastructure.' },
  { code: 'CYBER_ESSENTIALS', name: 'Cyber Essentials', version: '3.1', body: 'NCSC (UK)', controls: 5, color: '#43A047', category: 'Security', description: 'UK government-backed scheme for basic cyber hygiene.' },
  { code: 'NIST_800_53', name: 'NIST SP 800-53', version: 'Rev 5', body: 'NIST', controls: 191, color: '#E53935', category: 'Security', description: 'Comprehensive security and privacy controls for information systems.' },
  { code: 'NIST_CSF_2', name: 'NIST CSF 2.0', version: '2.0', body: 'NIST', controls: 78, color: '#F4511E', category: 'Security', description: 'Cybersecurity Framework with Govern, Identify, Protect, Detect, Respond, Recover.' },
  { code: 'PCI_DSS_4', name: 'PCI DSS v4.0', version: 'v4.0', body: 'PCI SSC', controls: 93, color: '#FF8F00', category: 'Security', description: 'Payment Card Industry standard for cardholder data protection.' },
  { code: 'ITIL_4', name: 'ITIL 4', version: '4', body: 'AXELOS / PeopleCert', controls: 34, color: '#00897B', category: 'Operational', description: 'IT service management best practices for digital transformation.' },
  { code: 'COBIT_2019', name: 'COBIT 2019', version: '2019', body: 'ISACA', controls: 40, color: '#5E35B1', category: 'Governance', description: 'Enterprise IT governance with 40 management objectives.' },
];

export default function FrameworksPage() {
  return (
    <div>
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Compliance Frameworks</h1>
          <p className="text-gray-500 mt-1">9 standards covering security, privacy, governance, and operations</p>
        </div>
        <button className="btn-primary">Adopt Framework</button>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
        {FRAMEWORKS.map(fw => (
          <a
            key={fw.code}
            href={`/frameworks/${fw.code}`}
            className="card hover:border-indigo-300 hover:shadow-md transition-all group"
          >
            <div className="flex items-start gap-3">
              <div className="h-10 w-10 rounded-lg flex items-center justify-center flex-shrink-0" style={{ backgroundColor: fw.color + '20' }}>
                <span className="text-sm font-bold" style={{ color: fw.color }}>{fw.code.substring(0, 2)}</span>
              </div>
              <div className="flex-1 min-w-0">
                <h3 className="font-semibold text-gray-900 group-hover:text-indigo-600 transition-colors">{fw.name}</h3>
                <p className="text-xs text-gray-500 mt-0.5">{fw.body} • v{fw.version}</p>
              </div>
              <span className={`badge ${fw.category === 'Security' ? 'badge-info' : fw.category === 'Privacy' ? 'badge-high' : fw.category === 'Governance' ? 'badge-medium' : 'badge-low'}`}>
                {fw.category}
              </span>
            </div>

            <p className="text-sm text-gray-600 mt-3">{fw.description}</p>

            <div className="flex items-center justify-between mt-4 pt-3 border-t border-gray-100">
              <span className="text-sm text-gray-500">{fw.controls} controls</span>
              <span className="text-xs font-medium text-indigo-600 group-hover:underline">View Controls →</span>
            </div>
          </a>
        ))}
      </div>
    </div>
  );
}
