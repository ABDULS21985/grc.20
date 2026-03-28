import type { Metadata } from 'next';
import './globals.css';

export const metadata: Metadata = {
  title: 'ComplianceForge — GRC Platform',
  description: 'Enterprise Governance, Risk & Compliance Management',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className="bg-gray-50 text-gray-900 antialiased">
        <div className="flex h-screen overflow-hidden">
          {/* Sidebar */}
          <aside className="hidden w-64 flex-shrink-0 border-r border-gray-200 bg-white lg:block">
            <div className="flex h-16 items-center gap-2 border-b border-gray-200 px-6">
              <div className="h-8 w-8 rounded-lg bg-indigo-600 flex items-center justify-center">
                <span className="text-white font-bold text-sm">CF</span>
              </div>
              <span className="font-semibold text-gray-900">ComplianceForge</span>
            </div>
            <nav className="mt-4 space-y-1 px-3">
              <NavItem href="/dashboard" icon="LayoutDashboard" label="Dashboard" />
              <NavItem href="/frameworks" icon="Shield" label="Frameworks" />
              <NavItem href="/risks" icon="AlertTriangle" label="Risk Register" />
              <NavItem href="/policies" icon="FileText" label="Policies" />
              <NavItem href="/audits" icon="ClipboardCheck" label="Audits" />
              <NavItem href="/incidents" icon="AlertCircle" label="Incidents" />
              <NavItem href="/vendors" icon="Building2" label="Vendors" />
              <NavItem href="/settings" icon="Settings" label="Settings" />
            </nav>
            <div className="absolute bottom-0 w-64 border-t border-gray-200 p-4">
              <div className="flex items-center gap-3">
                <div className="h-8 w-8 rounded-full bg-indigo-100 flex items-center justify-center">
                  <span className="text-indigo-600 text-xs font-medium">NU</span>
                </div>
                <div className="text-sm">
                  <p className="font-medium text-gray-900">Nemile</p>
                  <p className="text-gray-500 text-xs">Org Admin</p>
                </div>
              </div>
            </div>
          </aside>

          {/* Main content */}
          <main className="flex-1 overflow-y-auto">
            <div className="px-6 py-8 max-w-7xl mx-auto">
              {children}
            </div>
          </main>
        </div>
      </body>
    </html>
  );
}

function NavItem({ href, icon, label }: { href: string; icon: string; label: string }) {
  return (
    <a
      href={href}
      className="flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100 hover:text-gray-900 transition-colors"
    >
      <span className="text-gray-400">{icon.charAt(0)}</span>
      {label}
    </a>
  );
}
