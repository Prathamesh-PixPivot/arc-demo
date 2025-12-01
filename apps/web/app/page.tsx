export default function Home() {
    return (
        <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-purple-50 via-white to-purple-50">
            <div className="text-center space-y-6 p-8">
                <div className="inline-flex items-center justify-center w-20 h-20 bg-purple rounded-2xl shadow-elevated">
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        fill="none"
                        viewBox="0 0 24 24"
                        strokeWidth={2}
                        stroke="white"
                        className="w-12 h-12"
                    >
                        <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            d="M9 12.75 11.25 15 15 9.75m-3-7.036A11.959 11.959 0 0 1 3.598 6 11.99 11.99 0 0 0 3 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285Z"
                        />
                    </svg>
                </div>

                <div className="space-y-2">
                    <h1 className="text-5xl font-bold bg-gradient-to-r from-purple via-purple-hover to-purple-dark bg-clip-text text-transparent">
                        Arc Privacy Platform
                    </h1>
                    <p className="text-xl text-gray-600 max-w-2xl">
                        Enterprise-grade consent management and privacy compliance
                    </p>
                </div>

                <div className="flex gap-4 justify-center pt-4">
                    <a
                        href="/login"
                        className="btn-primary"
                    >
                        Get Started
                    </a>
                    <a
                        href="/docs"
                        className="btn-secondary"
                    >
                        Documentation
                    </a>
                </div>

                <div className="pt-8 border-t border-gray-200 mt-8">
                    <div className="flex items-center justify-center gap-8 text-sm text-gray-500">
                        <div className="flex items-center gap-2">
                            <div className="w-2 h-2 bg-success rounded-full"></div>
                            <span>WCAG 2.1 AA Compliant</span>
                        </div>
                        <div className="flex items-center gap-2">
                            <div className="w-2 h-2 bg-success rounded-full"></div>
                            <span>GDPR & DPDP Ready</span>
                        </div>
                        <div className="flex items-center gap-2">
                            <div className="w-2 h-2 bg-success rounded-full"></div>
                            <span>Enterprise Grade</span>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}
