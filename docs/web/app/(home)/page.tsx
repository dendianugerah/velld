"use client"

import { Button } from "@/components/ui/button"
import { Github, ArrowRight, Copy, Check } from "lucide-react"
import Image from "next/image"
import Link from "next/link"
import { useState } from "react"

export default function HomePage() {
  const [copiedCommand, setCopiedCommand] = useState<string | null>(null);

  const copyToClipboard = async (text: string, id: string) => {
    await navigator.clipboard.writeText(text);
    setCopiedCommand(id);
    setTimeout(() => setCopiedCommand(null), 2000);
  };

  return (
    <div className="min-h-screen bg-white dark:bg-black">
      <section className="container mx-auto px-6 pt-32 pb-32">
        <div className="max-w-4xl mx-auto text-center space-y-8">
          <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full border border-gray-200 dark:border-gray-800 text-sm mb-4">
            <span className="relative flex h-2 w-2">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-blue-400 opacity-75"></span>
              <span className="relative inline-flex rounded-full h-2 w-2 bg-blue-500"></span>
            </span>
            Open source database backup tool
          </div>

          <h1 className="text-[56px] md:text-[72px] lg:text-[96px] font-bold tracking-tight leading-[0.95]">
            Backup your
            <br />
            databases with
            <br />
            <span className="text-blue-600">ease</span>
          </h1>

          <p className="text-xl md:text-2xl text-gray-600 dark:text-gray-400 max-w-2xl mx-auto font-light">
            Automated scheduling, monitoring, and recovery for PostgreSQL, MySQL, and MongoDB.
          </p>

          <div className="flex flex-col sm:flex-row gap-4 justify-center items-center pt-4">
            <Link href="/docs" className="cursor-pointer">
              <Button size="lg" className="h-10 px-8 text-sm bg-blue-600 text-white hover:bg-blue-700 cursor-pointer shadow-sm">
                Get started
                <ArrowRight className="w-4 h-4 ml-2" />
              </Button>
            </Link>
            <a href="https://github.com/dendianugerah/velld" target="_blank" rel="noopener noreferrer" className="cursor-pointer">
              <Button variant="outline" size="lg" className="h-10 px-8 text-sm cursor-pointer shadow-sm">
                <Github className="w-4 h-4 mr-2" />
                View on GitHub
              </Button>
            </a>
          </div>

          <div className="pt-8 flex items-center justify-center gap-6 text-sm text-gray-500 dark:text-gray-500">
            <div className="flex items-center gap-2">
              <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
              </svg>
              Open source
            </div>
            <div className="flex items-center gap-2">
              <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/>
                <path d="M7 11V7a5 5 0 0 1 10 0v4"/>
              </svg>
              MIT Licensed
            </div>
            <div className="flex items-center gap-2">
              <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
                <circle cx="9" cy="7" r="4"/>
                <path d="M23 21v-2a4 4 0 0 0-3-3.87"/>
                <path d="M16 3.13a4 4 0 0 1 0 7.75"/>
              </svg>
              Community driven
            </div>
          </div>
        </div>
      </section>

      {/* Screenshot Section */}
      <section className="container mx-auto px-6 py-24">
        <div className="max-w-6xl mx-auto">
          <div className="rounded-2xl border border-gray-200 dark:border-gray-800 overflow-hidden shadow-2xl bg-gradient-to-b from-gray-50 to-white dark:from-gray-900 dark:to-black p-1">
            <Image
              src="/images/dashboard.png"
              alt="Velld Dashboard"
              width={1920}
              height={1080}
              className="w-full h-auto rounded-xl"
              priority
            />
          </div>
        </div>
      </section>

      <section className="container mx-auto px-6 py-32">
        <div className="max-w-6xl mx-auto">
          <div className="max-w-3xl mb-24">
            <h2 className="text-4xl md:text-5xl lg:text-6xl font-bold mb-6">
              Everything you need
            </h2>
            <p className="text-xl md:text-2xl text-gray-600 dark:text-gray-400">
              Built for everyone who needs reliable database backups
            </p>
          </div>

          <div className="grid md:grid-cols-2 gap-x-16 gap-y-16">
            <div className="space-y-4">
              <h3 className="text-2xl md:text-3xl font-bold">Multi-database support</h3>
              <p className="text-lg text-gray-600 dark:text-gray-400 leading-relaxed">
                Works seamlessly with PostgreSQL, MySQL, and MongoDB. More databases coming soon.
              </p>
            </div>

            <div className="space-y-4">
              <h3 className="text-2xl md:text-3xl font-bold">Automated scheduling</h3>
              <p className="text-lg text-gray-600 dark:text-gray-400 leading-relaxed">
                Set it and forget it. Schedule backups to run automatically on your preferred intervals.
              </p>
            </div>

            <div className="space-y-4">
              <h3 className="text-2xl md:text-3xl font-bold">Smart notifications</h3>
              <p className="text-lg text-gray-600 dark:text-gray-400 leading-relaxed">
                Get notified via email, webhooks, or dashboard when backups succeed or fail.
              </p>
            </div>

            <div className="space-y-4">
              <h3 className="text-2xl md:text-3xl font-bold">Secure & reliable</h3>
              <p className="text-lg text-gray-600 dark:text-gray-400 leading-relaxed">
                Your backups are secure with encrypted connections and reliable storage.
              </p>
            </div>

            <div className="space-y-4">
              <h3 className="text-2xl md:text-3xl font-bold">Fast recovery</h3>
              <p className="text-lg text-gray-600 dark:text-gray-400 leading-relaxed">
                Restore your databases quickly when you need them with just a few clicks.
              </p>
            </div>

            <div className="space-y-4">
              <h3 className="text-2xl md:text-3xl font-bold">Easy deployment</h3>
              <p className="text-lg text-gray-600 dark:text-gray-400 leading-relaxed">
                Deploy with Docker in minutes. No complicated setup or configuration required.
              </p>
            </div>
          </div>
        </div>
      </section>

      <section className="container mx-auto px-6 py-32 bg-gray-50 dark:bg-gray-950/50">
        <div className="max-w-6xl mx-auto">
          <div className="text-center mb-20">
            <h2 className="text-4xl md:text-5xl font-bold mb-4">
              See it in action
            </h2>
            <p className="text-xl text-gray-600 dark:text-gray-400">
              A beautiful interface for managing your database backups
            </p>
          </div>

          <div className="grid md:grid-cols-2 gap-12">
            <div className="space-y-6 group">
              <div className="rounded-2xl border border-gray-200 dark:border-gray-800 overflow-hidden transition-all duration-300 group-hover:border-gray-300 dark:group-hover:border-gray-700">
                <Image
                  src="/images/connections.png"
                  alt="Connection Management"
                  width={1200}
                  height={800}
                  className="w-full h-auto"
                />
              </div>
              <div className="space-y-2 px-2">
                <h3 className="text-2xl font-bold">Connection management</h3>
                <p className="text-lg text-gray-600 dark:text-gray-400">
                  Add and manage multiple database connections from a single dashboard.
                </p>
              </div>
            </div>

            <div className="space-y-6 group">
              <div className="rounded-2xl border border-gray-200 dark:border-gray-800 overflow-hidden transition-all duration-300 group-hover:border-gray-300 dark:group-hover:border-gray-700">
                <Image
                  src="/images/history.png"
                  alt="Backup History"
                  width={1200}
                  height={800}
                  className="w-full h-auto"
                />
              </div>
              <div className="space-y-2 px-2">
                <h3 className="text-2xl font-bold">Complete history</h3>
                <p className="text-lg text-gray-600 dark:text-gray-400">
                  Track all your backups with detailed logs and status information.
                </p>
              </div>
            </div>

            <div className="space-y-6 group md:col-span-2">
              <div className="rounded-2xl border border-gray-200 dark:border-gray-800 overflow-hidden transition-all duration-300 group-hover:border-gray-300 dark:group-hover:border-gray-700">
                <Image
                  src="/images/comparison.png"
                  alt="Backup Comparison"
                  width={1920}
                  height={1080}
                  className="w-full h-auto"
                />
              </div>
              <div className="space-y-2 px-2">
                <h3 className="text-2xl font-bold">Backup comparison</h3>
                <p className="text-lg text-gray-600 dark:text-gray-400">
                  Compare any two backups side-by-side with intelligent diff visualization.
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="container mx-auto px-6 py-32">
        <div className="max-w-4xl mx-auto">
          <div className="mb-16">
            <h2 className="text-4xl md:text-5xl lg:text-6xl font-bold mb-6">
              Get started in minutes
            </h2>
            <p className="text-xl md:text-2xl text-gray-600 dark:text-gray-400">
              Deploy Velld with Docker and start backing up your databases
            </p>
          </div>

          <div className="space-y-8">
            <div className="bg-black dark:bg-white/5 rounded-2xl p-6 md:p-10 border border-gray-800 dark:border-gray-700 overflow-x-auto relative group">
              <div className="font-mono text-xs sm:text-sm md:text-base space-y-4 min-w-0">
                <div className="flex items-start gap-3 sm:gap-4">
                  <span className="text-gray-500 dark:text-gray-500 select-none flex-shrink-0">$</span>
                  <div className="flex-1 min-w-0 flex items-center justify-between gap-4">
                    <div className="text-white dark:text-gray-100 break-all">git clone https://github.com/dendianugerah/velld.git</div>
                    <button
                      onClick={() => copyToClipboard('git clone https://github.com/dendianugerah/velld.git', 'clone')}
                      className="flex-shrink-0 p-2 rounded-md hover:bg-white/10 transition-colors opacity-0 group-hover:opacity-100"
                      title="Copy to clipboard"
                    >
                      {copiedCommand === 'clone' ? (
                        <Check className="w-4 h-4 text-green-400" />
                      ) : (
                        <Copy className="w-4 h-4 text-gray-400" />
                      )}
                    </button>
                  </div>
                </div>
                
                <div className="flex items-start gap-3 sm:gap-4">
                  <span className="text-gray-500 dark:text-gray-500 select-none flex-shrink-0">$</span>
                  <div className="flex-1 min-w-0 flex items-center justify-between gap-4">
                    <div className="text-white dark:text-gray-100 break-all">cd velld && docker compose up -d</div>
                    <button
                      onClick={() => copyToClipboard('cd velld && docker compose up -d', 'compose')}
                      className="flex-shrink-0 p-2 rounded-md hover:bg-white/10 transition-colors opacity-0 group-hover:opacity-100"
                      title="Copy to clipboard"
                    >
                      {copiedCommand === 'compose' ? (
                        <Check className="w-4 h-4 text-green-400" />
                      ) : (
                        <Copy className="w-4 h-4 text-gray-400" />
                      )}
                    </button>
                  </div>
                </div>
              </div>
            </div>

            <div className="grid sm:grid-cols-2 gap-6">
              <div className="p-6 rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-black">
                <div className="space-y-2">
                  <div className="text-sm font-medium text-gray-500 dark:text-gray-400">Web Interface</div>
                  <div className="text-xl sm:text-2xl font-bold break-all">localhost:3000</div>
                </div>
              </div>

              <div className="p-6 rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-black">
                <div className="space-y-2">
                  <div className="text-sm font-medium text-gray-500 dark:text-gray-400">API Endpoint</div>
                  <div className="text-xl sm:text-2xl font-bold break-all">localhost:8080</div>
                </div>
              </div>
            </div>

            {/* Next Steps */}
            <div className="pt-4">
              <Link href="/docs" className="inline-flex items-center gap-2 text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300 font-medium cursor-pointer">
                Read the documentation
                <ArrowRight className="w-4 h-4" />
              </Link>
            </div>
          </div>
        </div>
      </section>

      <section className="container mx-auto px-6 py-32 border-t border-gray-200 dark:border-gray-800">
        <div className="max-w-4xl mx-auto text-center space-y-8">
          <h2 className="text-4xl md:text-5xl lg:text-6xl font-bold">
            Ready to secure
            <br />
            your data?
          </h2>
          <p className="text-xl md:text-2xl text-gray-600 dark:text-gray-400 max-w-2xl mx-auto">
            Start backing up your databases with Velld
          </p>
        </div>
      </section>

      <footer className="container mx-auto px-6 py-12 border-t border-gray-200 dark:border-gray-800">
        <div className="max-w-6xl mx-auto">
          <div className="flex flex-col md:flex-row justify-between items-center gap-6">
            <div className="flex items-center gap-3">
              <Image
                src="/images/logo.png"
                alt="Velld Logo"
                width={36}
                height={36}
                className="rounded-lg"
              />
              <span className="text-base font-semibold bg-clip-text text-transparent bg-gradient-to-r from-blue-500 to-blue-600">
                Velld
              </span>
            </div>
            
            <div className="flex items-center gap-6 text-sm text-gray-600 dark:text-gray-400">
              <a href="https://github.com/dendianugerah/velld" target="_blank" rel="noopener noreferrer" className="hover:text-gray-900 dark:hover:text-gray-100">
                GitHub
              </a>
              <Link href="/docs" className="hover:text-gray-900 dark:hover:text-gray-100">
                Documentation
              </Link>
              <span>MIT License</span>
            </div>
          </div>
        </div>
      </footer>
    </div>
  )
}
