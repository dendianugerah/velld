"use client"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Github, Download, Play } from "lucide-react"
import Image from "next/image"
import Link  from "next/link"

export default function HomePage() {
  return (
    <div className="min-h-screen bg-background">
      {/* Hero Section */}
      <section className="container mx-auto px-4 py-16 md:py-24">
        <div className="text-center space-y-6 max-w-4xl mx-auto">
          <Badge variant="outline" className="mb-4">
            Open Source Database Backup Tool
          </Badge>

          <h1 className="text-4xl md:text-6xl font-bold tracking-tight">
            Secure Your Data with{" "}
            <span className="bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">Velld</span>
          </h1>

          <p className="text-xl text-muted-foreground max-w-2xl mx-auto leading-relaxed">
            A modern database backup management and automation tool. Schedule, manage, and monitor your database backups
            with ease, ensuring data security and effortless recovery.
          </p>

          <div className="flex flex-col sm:flex-row gap-4 justify-center items-center pt-6">
            <Button size="lg" className="gap-2">
              <Link href="/docs" className="flex items-center gap-2">
                <Play className="w-4 h-4" />
                Get Started
              </Link>
            </Button>
            <Button variant="outline" size="lg" className="gap-2">
              <a href="https://github.com/dendianugerah/velld"
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-2"
              >
                <Github className="w-4 h-4" />
                View on GitHub
              </a>
            </Button>
          </div>

          <div className="flex items-center justify-center gap-8 pt-8 text-sm text-muted-foreground">
            <div className="flex items-center gap-2">
              <div className="w-2 h-2 bg-green-500 rounded-full"></div>
              MIT Licensed
            </div>
            <div className="flex items-center gap-2">
              <div className="w-2 h-2 bg-blue-500 rounded-full"></div>
              Docker Ready
            </div>
            <div className="flex items-center gap-2">
              <div className="w-2 h-2 bg-purple-500 rounded-full"></div>
              Multi-Database
            </div>
          </div>
        </div>
      </section>

      {/* Supported Databases */}
      <section className="container mx-auto px-4 py-16">
        <div className="text-center mb-12">
          <h2 className="text-3xl font-bold mb-4">Supported Databases</h2>
          <p className="text-muted-foreground">Works with your favorite databases out of the box</p>
        </div>

        <div className="flex flex-wrap justify-center gap-8 max-w-3xl mx-auto">
          <div className="flex items-center gap-3 px-6 py-3">
            <div className="w-8 h-8 bg-blue-600 rounded flex items-center justify-center">
              <span className="text-white font-bold text-sm">P</span>
            </div>
            <span className="font-medium text-lg">PostgreSQL</span>
          </div>

          <div className="flex items-center gap-3 px-6 py-3">
            <div className="w-8 h-8 bg-orange-500 rounded flex items-center justify-center">
              <span className="text-white font-bold text-sm">M</span>
            </div>
            <span className="font-medium text-lg">MySQL</span>
          </div>

          <div className="flex items-center gap-3 px-6 py-3">
            <div className="w-8 h-8 bg-green-600 rounded flex items-center justify-center">
              <span className="text-white font-bold text-sm">M</span>
            </div>
            <span className="font-medium text-lg">MongoDB</span>
          </div>

          <div className="flex items-center gap-3 px-6 py-3 opacity-60">
            <div className="w-8 h-8 bg-gray-400 rounded flex items-center justify-center">
              <span className="text-white font-bold text-sm">+</span>
            </div>
            <span className="font-medium text-lg">More Coming Soon</span>
          </div>
        </div>
      </section>

      {/* Screenshots Section */}
      <section className="container mx-auto px-4 py-16">
        <div className="text-center mb-16">
          <h2 className="text-3xl font-bold mb-4">See Velld in Action</h2>
          <p className="text-muted-foreground">A clean, intuitive interface for managing your database backups</p>
        </div>

        <div className="grid md:grid-cols-3 gap-8 max-w-6xl mx-auto">
          <div className="space-y-4">
            <div className="aspect-video bg-gradient-to-br from-gray-50 to-gray-100 dark:from-gray-900 dark:to-gray-800 rounded-lg border flex items-center justify-center">
              <Image
                src="/placeholder.svg?height=200&width=300"
                alt="Dashboard Overview"
                width={300}
                height={200}
                className="rounded opacity-80"
              />
            </div>
            <div className="space-y-2">
              <h3 className="text-lg font-semibold">Dashboard Overview</h3>
              <p className="text-muted-foreground text-sm">
                Monitor backup statistics and recent activities at a glance
              </p>
            </div>
          </div>

          <div className="space-y-4">
            <div className="aspect-video bg-gradient-to-br from-gray-50 to-gray-100 dark:from-gray-900 dark:to-gray-800 rounded-lg border flex items-center justify-center">
              <Image
                src="/placeholder.svg?height=200&width=300"
                alt="Connection Management"
                width={300}
                height={200}
                className="rounded opacity-80"
              />
            </div>
            <div className="space-y-2">
              <h3 className="text-lg font-semibold">Connection Management</h3>
              <p className="text-muted-foreground text-sm">Easily manage multiple database connections in one place</p>
            </div>
          </div>

          <div className="space-y-4">
            <div className="aspect-video bg-gradient-to-br from-gray-50 to-gray-100 dark:from-gray-900 dark:to-gray-800 rounded-lg border flex items-center justify-center">
              <Image
                src="/placeholder.svg?height=200&width=300"
                alt="Backup History"
                width={300}
                height={200}
                className="rounded opacity-80"
              />
            </div>
            <div className="space-y-2">
              <h3 className="text-lg font-semibold">Backup History</h3>
              <p className="text-muted-foreground text-sm">View detailed backup history and logs for troubleshooting</p>
            </div>
          </div>
        </div>
      </section>

      {/* Quick Start Section */}
      <section className="container mx-auto px-4 py-16">
        <div className="max-w-4xl mx-auto">
          <div className="text-center mb-12">
            <h2 className="text-3xl font-bold mb-4">Quick Start</h2>
            <p className="text-muted-foreground">Get Velld up and running in minutes with Docker</p>
          </div>

          <div className="space-y-6">
            <div className="space-y-3">
              <div className="flex items-center gap-2 text-lg font-semibold">
                <Download className="w-5 h-5" />
                Installation
              </div>
              <p className="text-muted-foreground">Clone the repository and start with Docker Compose</p>
            </div>

            <div className="bg-muted/50 rounded-lg p-6 font-mono text-sm border">
              <div className="text-muted-foreground mb-2"># Clone the repository</div>
              <div className="mb-4">git clone https://github.com/dendianugerah/velld.git</div>
              <div className="text-muted-foreground mb-2"># Navigate and start</div>
              <div className="mb-1">cd velld</div>
              <div>docker compose up -d</div>
            </div>

            <div className="grid md:grid-cols-2 gap-6 pt-4">
              <div className="flex items-center gap-3 p-4 border rounded-lg">
                <div className="w-3 h-3 bg-blue-500 rounded-full"></div>
                <div>
                  <div className="font-medium">Web Interface</div>
                  <div className="text-sm text-muted-foreground">localhost:3000</div>
                </div>
              </div>

              <div className="flex items-center gap-3 p-4 border rounded-lg">
                <div className="w-3 h-3 bg-green-500 rounded-full"></div>
                <div>
                  <div className="font-medium">API Endpoint</div>
                  <div className="text-sm text-muted-foreground">localhost:8080</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Footer Section */}
      <section className="container mx-auto px-4 py-16 border-t">
        <div className="text-center space-y-6 max-w-3xl mx-auto">
          <p className="text-sm text-muted-foreground pt-4">Open source • MIT Licensed • Community driven</p>
        </div>
      </section>
    </div>
  )
}
