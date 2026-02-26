package services

// BuiltInTemplate represents a complete template definition with file templates
type BuiltInTemplate struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Description   string            `json:"description"`
	Category      string            `json:"category"`
	StackType     string            `json:"stack_type"` // svelte, go, fullstack
	FilesTemplate map[string]string `json:"files_template"`
	ConfigSchema  map[string]ConfigField `json:"config_schema"`
}

// ConfigField describes a configuration field for a template
type ConfigField struct {
	Type        string   `json:"type"` // string, number, boolean, select
	Label       string   `json:"label"`
	Description string   `json:"description"`
	Default     string   `json:"default"`
	Required    bool     `json:"required"`
	Options     []string `json:"options,omitempty"`
}

// builtInTemplates holds all built-in template definitions
var builtInTemplates = map[string]*BuiltInTemplate{
	"saas_dashboard":  saaDashboardTemplate(),
	"api_backend":     apiBackendTemplate(),
	"landing_page":    landingPageTemplate(),
	"crm_module":      crmModuleTemplate(),
	"task_manager":    taskManagerTemplate(),
}

// GetBuiltInTemplate returns a built-in template by name
func GetBuiltInTemplate(name string) (*BuiltInTemplate, bool) {
	t, ok := builtInTemplates[name]
	return t, ok
}

// GetAllBuiltInTemplates returns all built-in templates
func GetAllBuiltInTemplates() map[string]*BuiltInTemplate {
	return builtInTemplates
}

// --- SaaS Dashboard Template ---
func saaDashboardTemplate() *BuiltInTemplate {
	return &BuiltInTemplate{
		ID:          "saas_dashboard",
		Name:        "SaaS Dashboard",
		Description: "Full-featured SaaS dashboard with charts, user management, and analytics",
		Category:    "operations",
		StackType:   "svelte",
		ConfigSchema: map[string]ConfigField{
			"app_name":     {Type: "string", Label: "Application Name", Default: "My Dashboard", Required: true},
			"primary_color": {Type: "string", Label: "Primary Color", Default: "#3B82F6", Required: false},
			"chart_library": {Type: "select", Label: "Chart Library", Default: "chart.js", Options: []string{"chart.js", "d3", "echarts"}},
			"auth_enabled":  {Type: "boolean", Label: "Enable Authentication", Default: "true", Required: false},
		},
		FilesTemplate: map[string]string{
			"src/routes/+page.svelte": `<script lang="ts">
	import { onMount } from 'svelte';
	import StatsCard from '$lib/components/StatsCard.svelte';
	import RevenueChart from '$lib/components/RevenueChart.svelte';
	import UserTable from '$lib/components/UserTable.svelte';
	import ActivityFeed from '$lib/components/ActivityFeed.svelte';

	let stats = $state({
		totalUsers: 0,
		revenue: 0,
		activeSubscriptions: 0,
		churnRate: 0
	});

	onMount(async () => {
		// Fetch dashboard stats
		const response = await fetch('/api/dashboard/stats');
		if (response.ok) {
			stats = await response.json();
		}
	});
</script>

<svelte:head>
	<title>{{app_name}} - Dashboard</title>
</svelte:head>

<div class="min-h-screen bg-gray-50">
	<!-- Header -->
	<header class="bg-white border-b border-gray-200 px-6 py-4">
		<div class="flex items-center justify-between">
			<h1 class="text-2xl font-bold text-gray-900">{{app_name}}</h1>
			<div class="flex items-center gap-4">
				<button class="px-4 py-2 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700">
					Export Report
				</button>
			</div>
		</div>
	</header>

	<!-- Stats Grid -->
	<div class="p-6">
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
			<StatsCard title="Total Users" value={stats.totalUsers} trend="+12.5%" positive={true} />
			<StatsCard title="Revenue" value={'$' + stats.revenue.toLocaleString()} trend="+8.2%" positive={true} />
			<StatsCard title="Active Subscriptions" value={stats.activeSubscriptions} trend="+3.1%" positive={true} />
			<StatsCard title="Churn Rate" value={stats.churnRate + '%'} trend="-1.2%" positive={false} />
		</div>

		<!-- Charts -->
		<div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
			<RevenueChart />
			<ActivityFeed />
		</div>

		<!-- User Table -->
		<UserTable />
	</div>
</div>
`,
			"src/lib/components/StatsCard.svelte": `<script lang="ts">
	import { TrendingUp, TrendingDown } from 'lucide-svelte';

	interface Props {
		title: string;
		value: string | number;
		trend: string;
		positive: boolean;
	}

	let { title, value, trend, positive }: Props = $props();
</script>

<div class="bg-white rounded-xl border border-gray-200 p-6 hover:shadow-md transition-shadow">
	<div class="flex items-center justify-between mb-2">
		<span class="text-sm font-medium text-gray-600">{title}</span>
		<div class="flex items-center gap-1 text-sm {positive ? 'text-green-600' : 'text-red-600'}">
			{#if positive}
				<TrendingUp class="w-4 h-4" />
			{:else}
				<TrendingDown class="w-4 h-4" />
			{/if}
			<span>{trend}</span>
		</div>
	</div>
	<div class="text-3xl font-bold text-gray-900">{value}</div>
</div>
`,
			"src/lib/components/RevenueChart.svelte": `<script lang="ts">
	import { onMount } from 'svelte';

	let canvas: HTMLCanvasElement;
	let chartData = $state({
		labels: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun'],
		values: [4200, 5100, 4800, 6200, 7100, 8400]
	});

	onMount(async () => {
		// Initialize chart with {{chart_library}}
		const ctx = canvas.getContext('2d');
		if (ctx) {
			drawChart(ctx);
		}
	});

	function drawChart(ctx: CanvasRenderingContext2D) {
		const width = canvas.width;
		const height = canvas.height;
		const maxValue = Math.max(...chartData.values);
		const barWidth = width / chartData.values.length - 10;

		ctx.clearRect(0, 0, width, height);
		ctx.fillStyle = '{{primary_color}}';

		chartData.values.forEach((value, index) => {
			const barHeight = (value / maxValue) * (height - 40);
			const x = index * (barWidth + 10) + 5;
			const y = height - barHeight - 20;
			ctx.fillRect(x, y, barWidth, barHeight);
		});
	}
</script>

<div class="bg-white rounded-xl border border-gray-200 p-6">
	<h3 class="text-lg font-semibold text-gray-900 mb-4">Revenue Overview</h3>
	<canvas bind:this={canvas} width="500" height="300" class="w-full"></canvas>
</div>
`,
			"src/lib/components/UserTable.svelte": `<script lang="ts">
	import { Search, MoreVertical } from 'lucide-svelte';

	interface User {
		id: string;
		name: string;
		email: string;
		plan: string;
		status: 'active' | 'inactive' | 'trial';
		joinedAt: string;
	}

	let users = $state<User[]>([
		{ id: '1', name: 'Alice Johnson', email: 'alice@example.com', plan: 'Pro', status: 'active', joinedAt: '2024-01-15' },
		{ id: '2', name: 'Bob Smith', email: 'bob@example.com', plan: 'Basic', status: 'active', joinedAt: '2024-02-20' },
		{ id: '3', name: 'Carol Williams', email: 'carol@example.com', plan: 'Enterprise', status: 'trial', joinedAt: '2024-03-10' },
	]);

	let searchQuery = $state('');

	const filteredUsers = $derived(
		users.filter(u =>
			u.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
			u.email.toLowerCase().includes(searchQuery.toLowerCase())
		)
	);

	function getStatusColor(status: string): string {
		switch (status) {
			case 'active': return 'bg-green-100 text-green-700';
			case 'inactive': return 'bg-gray-100 text-gray-700';
			case 'trial': return 'bg-blue-100 text-blue-700';
			default: return 'bg-gray-100 text-gray-700';
		}
	}
</script>

<div class="bg-white rounded-xl border border-gray-200">
	<div class="flex items-center justify-between p-6 border-b border-gray-200">
		<h3 class="text-lg font-semibold text-gray-900">Users</h3>
		<div class="relative">
			<Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
			<input
				type="text"
				placeholder="Search users..."
				bind:value={searchQuery}
				class="pl-10 pr-4 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500"
			/>
		</div>
	</div>
	<div class="overflow-x-auto">
		<table class="w-full">
			<thead>
				<tr class="border-b border-gray-200">
					<th class="text-left text-xs font-medium text-gray-500 uppercase px-6 py-3">User</th>
					<th class="text-left text-xs font-medium text-gray-500 uppercase px-6 py-3">Plan</th>
					<th class="text-left text-xs font-medium text-gray-500 uppercase px-6 py-3">Status</th>
					<th class="text-left text-xs font-medium text-gray-500 uppercase px-6 py-3">Joined</th>
					<th class="text-right text-xs font-medium text-gray-500 uppercase px-6 py-3">Actions</th>
				</tr>
			</thead>
			<tbody>
				{#each filteredUsers as user (user.id)}
					<tr class="border-b border-gray-100 hover:bg-gray-50">
						<td class="px-6 py-4">
							<div>
								<div class="font-medium text-gray-900">{user.name}</div>
								<div class="text-sm text-gray-500">{user.email}</div>
							</div>
						</td>
						<td class="px-6 py-4 text-sm text-gray-700">{user.plan}</td>
						<td class="px-6 py-4">
							<span class="px-2.5 py-1 text-xs font-medium rounded-full {getStatusColor(user.status)}">
								{user.status}
							</span>
						</td>
						<td class="px-6 py-4 text-sm text-gray-500">{user.joinedAt}</td>
						<td class="px-6 py-4 text-right">
							<button class="p-1 text-gray-400 hover:text-gray-600 rounded">
								<MoreVertical class="w-4 h-4" />
							</button>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
</div>
`,
			"src/lib/components/ActivityFeed.svelte": `<script lang="ts">
	interface Activity {
		id: string;
		user: string;
		action: string;
		timestamp: string;
		type: 'signup' | 'upgrade' | 'payment' | 'cancel';
	}

	let activities = $state<Activity[]>([
		{ id: '1', user: 'Alice', action: 'upgraded to Pro plan', timestamp: '2 min ago', type: 'upgrade' },
		{ id: '2', user: 'Bob', action: 'made a payment of $49', timestamp: '15 min ago', type: 'payment' },
		{ id: '3', user: 'Carol', action: 'signed up for trial', timestamp: '1 hour ago', type: 'signup' },
		{ id: '4', user: 'Dave', action: 'cancelled subscription', timestamp: '2 hours ago', type: 'cancel' },
	]);

	function getTypeColor(type: string): string {
		switch (type) {
			case 'signup': return 'bg-blue-500';
			case 'upgrade': return 'bg-green-500';
			case 'payment': return 'bg-purple-500';
			case 'cancel': return 'bg-red-500';
			default: return 'bg-gray-500';
		}
	}
</script>

<div class="bg-white rounded-xl border border-gray-200 p-6">
	<h3 class="text-lg font-semibold text-gray-900 mb-4">Recent Activity</h3>
	<div class="space-y-4">
		{#each activities as activity (activity.id)}
			<div class="flex items-start gap-3">
				<div class="w-2 h-2 mt-2 rounded-full {getTypeColor(activity.type)}"></div>
				<div class="flex-1">
					<p class="text-sm text-gray-900">
						<span class="font-medium">{activity.user}</span> {activity.action}
					</p>
					<p class="text-xs text-gray-500">{activity.timestamp}</p>
				</div>
			</div>
		{/each}
	</div>
</div>
`,
			"src/app.css": `@tailwind base;
@tailwind components;
@tailwind utilities;

:root {
	--primary: {{primary_color}};
}
`,
			"package.json": `{
	"name": "{{app_name}}",
	"version": "1.0.0",
	"type": "module",
	"scripts": {
		"dev": "vite dev",
		"build": "vite build",
		"preview": "vite preview"
	},
	"devDependencies": {
		"@sveltejs/adapter-auto": "^3.0.0",
		"@sveltejs/kit": "^2.0.0",
		"svelte": "^5.0.0",
		"tailwindcss": "^3.4.0",
		"typescript": "^5.0.0",
		"vite": "^5.0.0"
	},
	"dependencies": {
		"lucide-svelte": "^0.300.0"
	}
}
`,
		},
	}
}

// --- API Backend Template ---
func apiBackendTemplate() *BuiltInTemplate {
	return &BuiltInTemplate{
		ID:          "api_backend",
		Name:        "API Backend",
		Description: "Go REST API with CRUD operations, authentication, and database integration",
		Category:    "operations",
		StackType:   "go",
		ConfigSchema: map[string]ConfigField{
			"app_name":      {Type: "string", Label: "Application Name", Default: "My API", Required: true},
			"module_name":   {Type: "string", Label: "Go Module Name", Default: "github.com/user/myapi", Required: true},
			"port":          {Type: "string", Label: "Server Port", Default: "8080", Required: false},
			"database_type": {Type: "select", Label: "Database", Default: "postgres", Options: []string{"postgres", "sqlite", "mysql"}},
			"auth_type":     {Type: "select", Label: "Authentication", Default: "jwt", Options: []string{"jwt", "session", "api_key"}},
		},
		FilesTemplate: map[string]string{
			"cmd/server/main.go": `package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"{{module_name}}/internal/handlers"
	"{{module_name}}/internal/middleware"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	slog.Info("starting {{app_name}}", "port", "{{port}}")

	mux := http.NewServeMux()

	// Register routes
	handlers.RegisterRoutes(mux)

	// Apply middleware
	handler := middleware.Chain(
		mux,
		middleware.Logger,
		middleware.CORS,
		middleware.Recovery,
	)

	srv := &http.Server{
		Addr:         ":{{port}}",
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	slog.Info("server started", "addr", srv.Addr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}

	slog.Info("server exited properly")
}
`,
			"internal/handlers/routes.go": `package handlers

import (
	"net/http"
)

// RegisterRoutes registers all API routes
func RegisterRoutes(mux *http.ServeMux) {
	// Health check
	mux.HandleFunc("GET /api/health", HealthCheck)

	// CRUD resources
	mux.HandleFunc("GET /api/items", ListItems)
	mux.HandleFunc("POST /api/items", CreateItem)
	mux.HandleFunc("GET /api/items/{id}", GetItem)
	mux.HandleFunc("PUT /api/items/{id}", UpdateItem)
	mux.HandleFunc("DELETE /api/items/{id}", DeleteItem)
}
`,
			"internal/handlers/health.go": `package handlers

import (
	"encoding/json"
	"net/http"
)

// HealthCheck returns the API health status
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":  "healthy",
		"service": "{{app_name}}",
		"version": "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
`,
			"internal/handlers/items.go": `package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Item represents a generic resource
type Item struct {
	ID          string    ` + "`" + `json:"id"` + "`" + `
	Name        string    ` + "`" + `json:"name"` + "`" + `
	Description string    ` + "`" + `json:"description"` + "`" + `
	Status      string    ` + "`" + `json:"status"` + "`" + `
	CreatedAt   time.Time ` + "`" + `json:"created_at"` + "`" + `
	UpdatedAt   time.Time ` + "`" + `json:"updated_at"` + "`" + `
}

// In-memory store (replace with {{database_type}} in production)
var (
	items   = make(map[string]Item)
	itemsMu sync.RWMutex
)

// ListItems returns all items
func ListItems(w http.ResponseWriter, r *http.Request) {
	itemsMu.RLock()
	defer itemsMu.RUnlock()

	result := make([]Item, 0, len(items))
	for _, item := range items {
		result = append(result, item)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"items": result,
		"total": len(result),
	})
}

// CreateItem creates a new item
func CreateItem(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string ` + "`" + `json:"name"` + "`" + `
		Description string ` + "`" + `json:"description"` + "`" + `
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}

	item := Item{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	itemsMu.Lock()
	items[item.ID] = item
	itemsMu.Unlock()

	slog.Info("item created", "id", item.ID, "name", item.Name)
	respondJSON(w, http.StatusCreated, item)
}

// GetItem returns a single item
func GetItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	itemsMu.RLock()
	item, exists := items[id]
	itemsMu.RUnlock()

	if !exists {
		respondError(w, http.StatusNotFound, "item not found")
		return
	}

	respondJSON(w, http.StatusOK, item)
}

// UpdateItem updates an existing item
func UpdateItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	itemsMu.RLock()
	item, exists := items[id]
	itemsMu.RUnlock()

	if !exists {
		respondError(w, http.StatusNotFound, "item not found")
		return
	}

	var req struct {
		Name        *string ` + "`" + `json:"name"` + "`" + `
		Description *string ` + "`" + `json:"description"` + "`" + `
		Status      *string ` + "`" + `json:"status"` + "`" + `
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name != nil {
		item.Name = *req.Name
	}
	if req.Description != nil {
		item.Description = *req.Description
	}
	if req.Status != nil {
		item.Status = *req.Status
	}
	item.UpdatedAt = time.Now()

	itemsMu.Lock()
	items[id] = item
	itemsMu.Unlock()

	slog.Info("item updated", "id", item.ID)
	respondJSON(w, http.StatusOK, item)
}

// DeleteItem deletes an item
func DeleteItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	itemsMu.Lock()
	_, exists := items[id]
	if exists {
		delete(items, id)
	}
	itemsMu.Unlock()

	if !exists {
		respondError(w, http.StatusNotFound, "item not found")
		return
	}

	slog.Info("item deleted", "id", id)
	w.WriteHeader(http.StatusNoContent)
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
`,
			"internal/middleware/middleware.go": `package middleware

import (
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

// Middleware is a function that wraps an http.Handler
type Middleware func(http.Handler) http.Handler

// Chain applies middlewares in order
func Chain(handler http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// Logger logs HTTP requests
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(start).String(),
		)
	})
}

// CORS adds CORS headers with origin validation
func CORS(next http.Handler) http.Handler {
	allowed := os.Getenv("ALLOWED_ORIGINS")
	var origins []string
	if allowed != "" {
		for _, o := range strings.Split(allowed, ",") {
			origins = append(origins, strings.TrimSpace(o))
		}
	} else {
		// Default: localhost only (no wildcard)
		origins = []string{"http://localhost:5173", "http://localhost:3000"}

		// PRODUCTION GUARD: Warn if using default origins in production
		env := os.Getenv("ENVIRONMENT")
		if env == "production" || env == "prod" || os.Getenv("ENV") == "production" {
			slog.Error("SECURITY WARNING: Using default CORS origins in production. Set ALLOWED_ORIGINS environment variable to explicit domains.")
			// In strict mode, you could panic here:
			// panic("CORS MISCONFIGURATION: ALLOWED_ORIGINS must be set in production")
		}
	}

	originSet := make(map[string]bool, len(origins))
	for _, o := range origins {
		// Normalize: lowercase, trim trailing slashes
		normalized := strings.TrimRight(strings.ToLower(strings.TrimSpace(o)), "/")
		originSet[normalized] = true
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqOrigin := r.Header.Get("Origin")
		// Normalize incoming origin for case-insensitive comparison
		normalizedReq := strings.TrimRight(strings.ToLower(strings.TrimSpace(reqOrigin)), "/")
		if originSet[normalizedReq] {
			w.Header().Set("Access-Control-Allow-Origin", reqOrigin) // Reflect original
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Vary", "Origin")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Recovery recovers from panics
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered", "error", err, "path", r.URL.Path)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
`,
			"go.mod": `module {{module_name}}

go 1.22

require (
	github.com/google/uuid v1.6.0
)
`,
			"Dockerfile": `FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/server .

EXPOSE {{port}}
CMD ["./server"]
`,
		},
	}
}

// --- Landing Page Template ---
func landingPageTemplate() *BuiltInTemplate {
	return &BuiltInTemplate{
		ID:          "landing_page",
		Name:        "Landing Page",
		Description: "Modern landing page with hero section, features, pricing, and contact form",
		Category:    "marketing",
		StackType:   "svelte",
		ConfigSchema: map[string]ConfigField{
			"app_name":       {Type: "string", Label: "Product Name", Default: "My Product", Required: true},
			"tagline":        {Type: "string", Label: "Tagline", Default: "The best solution for your needs", Required: false},
			"primary_color":  {Type: "string", Label: "Primary Color", Default: "#6366F1", Required: false},
			"cta_text":       {Type: "string", Label: "CTA Button Text", Default: "Get Started Free", Required: false},
			"pricing_enabled": {Type: "boolean", Label: "Show Pricing Section", Default: "true", Required: false},
		},
		FilesTemplate: map[string]string{
			"src/routes/+page.svelte": `<script lang="ts">
	import Hero from '$lib/components/Hero.svelte';
	import Features from '$lib/components/Features.svelte';
	import Pricing from '$lib/components/Pricing.svelte';
	import ContactForm from '$lib/components/ContactForm.svelte';
	import Footer from '$lib/components/Footer.svelte';
</script>

<svelte:head>
	<title>{{app_name}} - {{tagline}}</title>
	<meta name="description" content="{{tagline}}" />
</svelte:head>

<div class="min-h-screen">
	<Hero />
	<Features />
	<Pricing />
	<ContactForm />
	<Footer />
</div>
`,
			"src/lib/components/Hero.svelte": `<script lang="ts">
	import { ArrowRight } from 'lucide-svelte';
</script>

<section class="relative bg-gradient-to-br from-indigo-50 via-white to-purple-50 pt-20 pb-32">
	<div class="max-w-7xl mx-auto px-6 text-center">
		<h1 class="text-5xl md:text-7xl font-bold text-gray-900 mb-6 tracking-tight">
			{{app_name}}
		</h1>
		<p class="text-xl md:text-2xl text-gray-600 mb-10 max-w-3xl mx-auto">
			{{tagline}}
		</p>
		<div class="flex flex-col sm:flex-row items-center justify-center gap-4">
			<a
				href="#pricing"
				class="px-8 py-4 text-lg font-semibold text-white rounded-xl shadow-lg hover:shadow-xl transition-all"
				style="background-color: {{primary_color}}"
			>
				{{cta_text}}
				<ArrowRight class="w-5 h-5 inline ml-2" />
			</a>
			<a href="#features" class="px-8 py-4 text-lg font-semibold text-gray-700 bg-white border-2 border-gray-200 rounded-xl hover:border-gray-300">
				Learn More
			</a>
		</div>
	</div>
</section>
`,
			"src/lib/components/Features.svelte": `<script lang="ts">
	import { Zap, Shield, BarChart3, Users } from 'lucide-svelte';

	const features = [
		{
			icon: Zap,
			title: 'Lightning Fast',
			description: 'Built for performance with optimized workflows that save you hours every week.'
		},
		{
			icon: Shield,
			title: 'Enterprise Security',
			description: 'Bank-grade encryption and compliance standards to keep your data safe.'
		},
		{
			icon: BarChart3,
			title: 'Advanced Analytics',
			description: 'Deep insights into your business with real-time dashboards and reports.'
		},
		{
			icon: Users,
			title: 'Team Collaboration',
			description: 'Work together seamlessly with real-time editing and shared workspaces.'
		}
	];
</script>

<section id="features" class="py-24 bg-white">
	<div class="max-w-7xl mx-auto px-6">
		<div class="text-center mb-16">
			<h2 class="text-3xl md:text-4xl font-bold text-gray-900 mb-4">
				Everything you need
			</h2>
			<p class="text-lg text-gray-600 max-w-2xl mx-auto">
				Powerful features designed to help your team succeed.
			</p>
		</div>
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8">
			{#each features as feature}
				<div class="p-6 rounded-2xl border border-gray-200 hover:border-indigo-200 hover:shadow-lg transition-all">
					<div class="w-12 h-12 rounded-xl flex items-center justify-center mb-4" style="background-color: {{primary_color}}20">
						<svelte:component this={feature.icon} class="w-6 h-6" style="color: {{primary_color}}" />
					</div>
					<h3 class="text-lg font-semibold text-gray-900 mb-2">{feature.title}</h3>
					<p class="text-gray-600">{feature.description}</p>
				</div>
			{/each}
		</div>
	</div>
</section>
`,
			"src/lib/components/Pricing.svelte": `<script lang="ts">
	import { Check } from 'lucide-svelte';

	const plans = [
		{
			name: 'Starter',
			price: '19',
			description: 'Perfect for individuals',
			features: ['5 projects', '1 GB storage', 'Email support', 'Basic analytics']
		},
		{
			name: 'Pro',
			price: '49',
			description: 'Best for growing teams',
			features: ['Unlimited projects', '10 GB storage', 'Priority support', 'Advanced analytics', 'API access', 'Custom integrations'],
			popular: true
		},
		{
			name: 'Enterprise',
			price: '99',
			description: 'For large organizations',
			features: ['Everything in Pro', 'Unlimited storage', '24/7 phone support', 'Custom contracts', 'SLA guarantee', 'Dedicated manager']
		}
	];
</script>

<section id="pricing" class="py-24 bg-gray-50">
	<div class="max-w-7xl mx-auto px-6">
		<div class="text-center mb-16">
			<h2 class="text-3xl md:text-4xl font-bold text-gray-900 mb-4">Simple Pricing</h2>
			<p class="text-lg text-gray-600">Choose the plan that works for you.</p>
		</div>
		<div class="grid grid-cols-1 md:grid-cols-3 gap-8 max-w-5xl mx-auto">
			{#each plans as plan}
				<div class="relative bg-white rounded-2xl border-2 p-8 {plan.popular ? 'border-indigo-500 shadow-xl' : 'border-gray-200'}">
					{#if plan.popular}
						<div class="absolute -top-4 left-1/2 -translate-x-1/2 px-4 py-1 text-sm font-semibold text-white rounded-full" style="background-color: {{primary_color}}">
							Most Popular
						</div>
					{/if}
					<h3 class="text-xl font-bold text-gray-900 mb-2">{plan.name}</h3>
					<p class="text-gray-600 mb-4">{plan.description}</p>
					<div class="mb-6">
						<span class="text-4xl font-bold text-gray-900">${plan.price}</span>
						<span class="text-gray-500">/month</span>
					</div>
					<ul class="space-y-3 mb-8">
						{#each plan.features as feature}
							<li class="flex items-center gap-2 text-sm text-gray-700">
								<Check class="w-4 h-4 text-green-500 flex-shrink-0" />
								<span>{feature}</span>
							</li>
						{/each}
					</ul>
					<button
						class="w-full py-3 rounded-xl font-semibold transition-colors {plan.popular ? 'text-white' : 'text-gray-700 bg-gray-100 hover:bg-gray-200'}"
						style={plan.popular ? 'background-color: {{primary_color}}' : ''}
					>
						{{cta_text}}
					</button>
				</div>
			{/each}
		</div>
	</div>
</section>
`,
			"src/lib/components/ContactForm.svelte": `<script lang="ts">
	let name = $state('');
	let email = $state('');
	let message = $state('');
	let submitted = $state(false);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		// Submit form logic here
		submitted = true;
		setTimeout(() => { submitted = false; }, 3000);
	}
</script>

<section id="contact" class="py-24 bg-white">
	<div class="max-w-3xl mx-auto px-6">
		<div class="text-center mb-12">
			<h2 class="text-3xl font-bold text-gray-900 mb-4">Get in Touch</h2>
			<p class="text-gray-600">Have questions? We would love to hear from you.</p>
		</div>
		{#if submitted}
			<div class="p-6 bg-green-50 border border-green-200 rounded-xl text-center">
				<p class="text-green-700 font-medium">Thank you! We will get back to you soon.</p>
			</div>
		{:else}
			<form onsubmit={handleSubmit} class="space-y-6">
				<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
					<div>
						<label for="name" class="block text-sm font-medium text-gray-700 mb-2">Name</label>
						<input id="name" type="text" bind:value={name} required class="w-full px-4 py-3 border border-gray-300 rounded-xl focus:ring-2 focus:ring-indigo-500" />
					</div>
					<div>
						<label for="email" class="block text-sm font-medium text-gray-700 mb-2">Email</label>
						<input id="email" type="email" bind:value={email} required class="w-full px-4 py-3 border border-gray-300 rounded-xl focus:ring-2 focus:ring-indigo-500" />
					</div>
				</div>
				<div>
					<label for="message" class="block text-sm font-medium text-gray-700 mb-2">Message</label>
					<textarea id="message" bind:value={message} rows="4" required class="w-full px-4 py-3 border border-gray-300 rounded-xl focus:ring-2 focus:ring-indigo-500"></textarea>
				</div>
				<button type="submit" class="w-full py-4 text-white font-semibold rounded-xl" style="background-color: {{primary_color}}">
					Send Message
				</button>
			</form>
		{/if}
	</div>
</section>
`,
			"src/lib/components/Footer.svelte": `<footer class="bg-gray-900 text-gray-400 py-12">
	<div class="max-w-7xl mx-auto px-6">
		<div class="flex flex-col md:flex-row items-center justify-between">
			<div class="text-lg font-bold text-white mb-4 md:mb-0">{{app_name}}</div>
			<div class="flex items-center gap-6 text-sm">
				<a href="#features" class="hover:text-white transition-colors">Features</a>
				<a href="#pricing" class="hover:text-white transition-colors">Pricing</a>
				<a href="#contact" class="hover:text-white transition-colors">Contact</a>
			</div>
		</div>
		<div class="mt-8 pt-8 border-t border-gray-800 text-center text-sm">
			<p>Generated with BusinessOS Template System</p>
		</div>
	</div>
</footer>
`,
			"package.json": `{
	"name": "{{app_name}}",
	"version": "1.0.0",
	"type": "module",
	"scripts": {
		"dev": "vite dev",
		"build": "vite build",
		"preview": "vite preview"
	},
	"devDependencies": {
		"@sveltejs/adapter-auto": "^3.0.0",
		"@sveltejs/kit": "^2.0.0",
		"svelte": "^5.0.0",
		"tailwindcss": "^3.4.0",
		"typescript": "^5.0.0",
		"vite": "^5.0.0"
	},
	"dependencies": {
		"lucide-svelte": "^0.300.0"
	}
}
`,
		},
	}
}

// --- CRM Module Template ---
func crmModuleTemplate() *BuiltInTemplate {
	return &BuiltInTemplate{
		ID:          "crm_module",
		Name:        "CRM Module",
		Description: "Contact and deal management with pipeline view, activity tracking, and reporting",
		Category:    "crm",
		StackType:   "svelte",
		ConfigSchema: map[string]ConfigField{
			"app_name":       {Type: "string", Label: "CRM Name", Default: "My CRM", Required: true},
			"primary_color":  {Type: "string", Label: "Primary Color", Default: "#10B981", Required: false},
			"pipeline_stages": {Type: "string", Label: "Pipeline Stages (comma-separated)", Default: "Lead,Qualified,Proposal,Negotiation,Won,Lost", Required: false},
			"currency":       {Type: "select", Label: "Currency", Default: "USD", Options: []string{"USD", "EUR", "GBP", "BRL"}},
		},
		FilesTemplate: map[string]string{
			"src/routes/+page.svelte": `<script lang="ts">
	import ContactList from '$lib/components/ContactList.svelte';
	import DealPipeline from '$lib/components/DealPipeline.svelte';
	import CRMStats from '$lib/components/CRMStats.svelte';

	let activeView = $state<'contacts' | 'pipeline' | 'stats'>('pipeline');
</script>

<svelte:head>
	<title>{{app_name}}</title>
</svelte:head>

<div class="min-h-screen bg-gray-50">
	<!-- Header -->
	<header class="bg-white border-b border-gray-200 px-6 py-4">
		<div class="flex items-center justify-between">
			<h1 class="text-2xl font-bold text-gray-900">{{app_name}}</h1>
			<div class="flex items-center gap-2">
				<button
					onclick={() => activeView = 'pipeline'}
					class="px-4 py-2 text-sm rounded-lg transition-colors {activeView === 'pipeline' ? 'bg-emerald-100 text-emerald-700' : 'text-gray-600 hover:bg-gray-100'}"
				>Pipeline</button>
				<button
					onclick={() => activeView = 'contacts'}
					class="px-4 py-2 text-sm rounded-lg transition-colors {activeView === 'contacts' ? 'bg-emerald-100 text-emerald-700' : 'text-gray-600 hover:bg-gray-100'}"
				>Contacts</button>
				<button
					onclick={() => activeView = 'stats'}
					class="px-4 py-2 text-sm rounded-lg transition-colors {activeView === 'stats' ? 'bg-emerald-100 text-emerald-700' : 'text-gray-600 hover:bg-gray-100'}"
				>Stats</button>
			</div>
		</div>
	</header>

	<!-- Content -->
	<div class="p-6">
		{#if activeView === 'pipeline'}
			<DealPipeline />
		{:else if activeView === 'contacts'}
			<ContactList />
		{:else}
			<CRMStats />
		{/if}
	</div>
</div>
`,
			"src/lib/components/DealPipeline.svelte": `<script lang="ts">
	interface Deal {
		id: string;
		name: string;
		company: string;
		value: number;
		stage: string;
		probability: number;
	}

	const stages = '{{pipeline_stages}}'.split(',').map(s => s.trim());
	const currency = '{{currency}}';

	let deals = $state<Deal[]>([
		{ id: '1', name: 'Enterprise License', company: 'Acme Corp', value: 50000, stage: 'Proposal', probability: 60 },
		{ id: '2', name: 'Consulting Package', company: 'Tech Inc', value: 25000, stage: 'Qualified', probability: 40 },
		{ id: '3', name: 'Annual Subscription', company: 'Global Ltd', value: 12000, stage: 'Negotiation', probability: 80 },
		{ id: '4', name: 'Platform Migration', company: 'StartupXYZ', value: 75000, stage: 'Lead', probability: 20 },
	]);

	function getDealsForStage(stage: string): Deal[] {
		return deals.filter(d => d.stage === stage);
	}

	function getStageTotal(stage: string): number {
		return getDealsForStage(stage).reduce((sum, d) => sum + d.value, 0);
	}

	function formatCurrency(value: number): string {
		return new Intl.NumberFormat('en-US', { style: 'currency', currency: currency }).format(value);
	}
</script>

<div class="flex gap-4 overflow-x-auto pb-4">
	{#each stages as stage}
		<div class="flex-shrink-0 w-72 bg-gray-100 rounded-xl p-4">
			<div class="flex items-center justify-between mb-4">
				<h3 class="font-semibold text-gray-900">{stage}</h3>
				<span class="text-xs text-gray-500">{formatCurrency(getStageTotal(stage))}</span>
			</div>
			<div class="space-y-3">
				{#each getDealsForStage(stage) as deal (deal.id)}
					<div class="bg-white rounded-lg p-4 border border-gray-200 shadow-sm hover:shadow-md transition-shadow cursor-pointer">
						<h4 class="font-medium text-gray-900 text-sm mb-1">{deal.name}</h4>
						<p class="text-xs text-gray-500 mb-2">{deal.company}</p>
						<div class="flex items-center justify-between">
							<span class="text-sm font-semibold" style="color: {{primary_color}}">{formatCurrency(deal.value)}</span>
							<span class="text-xs text-gray-400">{deal.probability}%</span>
						</div>
					</div>
				{/each}
				{#if getDealsForStage(stage).length === 0}
					<div class="text-center py-4 text-sm text-gray-400">No deals</div>
				{/if}
			</div>
		</div>
	{/each}
</div>
`,
			"src/lib/components/ContactList.svelte": `<script lang="ts">
	import { Search, Plus, Mail, Phone } from 'lucide-svelte';

	interface Contact {
		id: string;
		name: string;
		email: string;
		phone: string;
		company: string;
		status: 'active' | 'inactive' | 'prospect';
		lastContact: string;
	}

	let contacts = $state<Contact[]>([
		{ id: '1', name: 'John Doe', email: 'john@acme.com', phone: '+1-555-0101', company: 'Acme Corp', status: 'active', lastContact: '2 days ago' },
		{ id: '2', name: 'Jane Smith', email: 'jane@tech.com', phone: '+1-555-0102', company: 'Tech Inc', status: 'active', lastContact: '1 week ago' },
		{ id: '3', name: 'Bob Wilson', email: 'bob@global.com', phone: '+1-555-0103', company: 'Global Ltd', status: 'prospect', lastContact: '3 days ago' },
	]);

	let searchQuery = $state('');

	const filteredContacts = $derived(
		contacts.filter(c =>
			c.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
			c.company.toLowerCase().includes(searchQuery.toLowerCase()) ||
			c.email.toLowerCase().includes(searchQuery.toLowerCase())
		)
	);
</script>

<div class="bg-white rounded-xl border border-gray-200">
	<div class="flex items-center justify-between p-6 border-b border-gray-200">
		<div class="flex items-center gap-4">
			<div class="relative">
				<Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
				<input
					type="text"
					placeholder="Search contacts..."
					bind:value={searchQuery}
					class="pl-10 pr-4 py-2 border border-gray-300 rounded-lg text-sm"
				/>
			</div>
		</div>
		<button class="flex items-center gap-2 px-4 py-2 text-sm text-white rounded-lg" style="background-color: {{primary_color}}">
			<Plus class="w-4 h-4" />
			Add Contact
		</button>
	</div>
	<div class="divide-y divide-gray-100">
		{#each filteredContacts as contact (contact.id)}
			<div class="flex items-center justify-between p-4 hover:bg-gray-50">
				<div class="flex items-center gap-4">
					<div class="w-10 h-10 rounded-full bg-gray-200 flex items-center justify-center text-sm font-medium text-gray-600">
						{contact.name.split(' ').map(n => n[0]).join('')}
					</div>
					<div>
						<div class="font-medium text-gray-900">{contact.name}</div>
						<div class="text-sm text-gray-500">{contact.company}</div>
					</div>
				</div>
				<div class="flex items-center gap-4">
					<span class="text-xs text-gray-400">Last: {contact.lastContact}</span>
					<button class="p-2 text-gray-400 hover:text-gray-600 rounded-lg hover:bg-gray-100">
						<Mail class="w-4 h-4" />
					</button>
					<button class="p-2 text-gray-400 hover:text-gray-600 rounded-lg hover:bg-gray-100">
						<Phone class="w-4 h-4" />
					</button>
				</div>
			</div>
		{/each}
	</div>
</div>
`,
			"src/lib/components/CRMStats.svelte": `<script lang="ts">
	const stats = [
		{ label: 'Total Contacts', value: '248', change: '+12%' },
		{ label: 'Active Deals', value: '34', change: '+5%' },
		{ label: 'Pipeline Value', value: '$1.2M', change: '+23%' },
		{ label: 'Win Rate', value: '68%', change: '+4%' },
	];
</script>

<div class="space-y-6">
	<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
		{#each stats as stat}
			<div class="bg-white rounded-xl border border-gray-200 p-6">
				<div class="text-sm text-gray-500 mb-1">{stat.label}</div>
				<div class="text-3xl font-bold text-gray-900">{stat.value}</div>
				<div class="text-sm text-green-600 mt-1">{stat.change} from last month</div>
			</div>
		{/each}
	</div>
</div>
`,
			"package.json": `{
	"name": "{{app_name}}",
	"version": "1.0.0",
	"type": "module",
	"scripts": {
		"dev": "vite dev",
		"build": "vite build",
		"preview": "vite preview"
	},
	"devDependencies": {
		"@sveltejs/adapter-auto": "^3.0.0",
		"@sveltejs/kit": "^2.0.0",
		"svelte": "^5.0.0",
		"tailwindcss": "^3.4.0",
		"typescript": "^5.0.0",
		"vite": "^5.0.0"
	},
	"dependencies": {
		"lucide-svelte": "^0.300.0"
	}
}
`,
		},
	}
}

// --- Task Manager Template ---
func taskManagerTemplate() *BuiltInTemplate {
	return &BuiltInTemplate{
		ID:          "task_manager",
		Name:        "Task Manager",
		Description: "Kanban-style task board with drag-and-drop, labels, due dates, and team assignment",
		Category:    "project_management",
		StackType:   "svelte",
		ConfigSchema: map[string]ConfigField{
			"app_name":      {Type: "string", Label: "App Name", Default: "Task Board", Required: true},
			"primary_color": {Type: "string", Label: "Primary Color", Default: "#8B5CF6", Required: false},
			"columns":       {Type: "string", Label: "Board Columns (comma-separated)", Default: "Backlog,To Do,In Progress,Review,Done", Required: false},
			"labels":        {Type: "string", Label: "Labels (comma-separated)", Default: "Bug,Feature,Enhancement,Urgent", Required: false},
		},
		FilesTemplate: map[string]string{
			"src/routes/+page.svelte": `<script lang="ts">
	import KanbanBoard from '$lib/components/KanbanBoard.svelte';
	import TaskModal from '$lib/components/TaskModal.svelte';
	import BoardHeader from '$lib/components/BoardHeader.svelte';

	let showModal = $state(false);
	let selectedTask = $state<any>(null);

	function handleAddTask() {
		selectedTask = null;
		showModal = true;
	}

	function handleEditTask(task: any) {
		selectedTask = task;
		showModal = true;
	}

	function handleCloseModal() {
		showModal = false;
		selectedTask = null;
	}
</script>

<svelte:head>
	<title>{{app_name}}</title>
</svelte:head>

<div class="min-h-screen bg-gray-50 flex flex-col">
	<BoardHeader onAddTask={handleAddTask} />
	<div class="flex-1 overflow-hidden">
		<KanbanBoard onEditTask={handleEditTask} />
	</div>
	{#if showModal}
		<TaskModal task={selectedTask} onClose={handleCloseModal} />
	{/if}
</div>
`,
			"src/lib/components/BoardHeader.svelte": `<script lang="ts">
	import { Plus, Search, Filter } from 'lucide-svelte';

	interface Props {
		onAddTask: () => void;
	}

	let { onAddTask }: Props = $props();
	let searchQuery = $state('');
</script>

<header class="bg-white border-b border-gray-200 px-6 py-4">
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-bold text-gray-900">{{app_name}}</h1>
		<div class="flex items-center gap-3">
			<div class="relative">
				<Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
				<input
					type="text"
					placeholder="Search tasks..."
					bind:value={searchQuery}
					class="pl-10 pr-4 py-2 border border-gray-300 rounded-lg text-sm w-64"
				/>
			</div>
			<button class="flex items-center gap-2 px-4 py-2 text-white text-sm font-medium rounded-lg" style="background-color: {{primary_color}}" onclick={onAddTask}>
				<Plus class="w-4 h-4" />
				Add Task
			</button>
		</div>
	</div>
</header>
`,
			"src/lib/components/KanbanBoard.svelte": `<script lang="ts">
	import { GripVertical, Clock, Tag } from 'lucide-svelte';

	interface Task {
		id: string;
		title: string;
		description: string;
		column: string;
		labels: string[];
		assignee: string;
		dueDate: string | null;
		priority: 'low' | 'medium' | 'high';
	}

	interface Props {
		onEditTask: (task: Task) => void;
	}

	let { onEditTask }: Props = $props();

	const columns = '{{columns}}'.split(',').map(c => c.trim());
	const availableLabels = '{{labels}}'.split(',').map(l => l.trim());

	let tasks = $state<Task[]>([
		{ id: '1', title: 'Design user dashboard', description: 'Create wireframes and mockups', column: 'In Progress', labels: ['Feature'], assignee: 'Alice', dueDate: '2025-02-15', priority: 'high' },
		{ id: '2', title: 'Fix login redirect', description: 'Users are not redirected after login', column: 'To Do', labels: ['Bug', 'Urgent'], assignee: 'Bob', dueDate: '2025-02-10', priority: 'high' },
		{ id: '3', title: 'Add dark mode support', description: 'Implement theme switching', column: 'Backlog', labels: ['Enhancement'], assignee: 'Carol', dueDate: null, priority: 'low' },
		{ id: '4', title: 'API rate limiting', description: 'Add rate limiting to public endpoints', column: 'Review', labels: ['Feature'], assignee: 'Dave', dueDate: '2025-02-12', priority: 'medium' },
		{ id: '5', title: 'Update documentation', description: 'Add API docs for v2 endpoints', column: 'Done', labels: ['Enhancement'], assignee: 'Eve', dueDate: null, priority: 'low' },
	]);

	function getTasksForColumn(column: string): Task[] {
		return tasks.filter(t => t.column === column);
	}

	function getLabelColor(label: string): string {
		const colors: Record<string, string> = {
			'Bug': 'bg-red-100 text-red-700',
			'Feature': 'bg-blue-100 text-blue-700',
			'Enhancement': 'bg-green-100 text-green-700',
			'Urgent': 'bg-orange-100 text-orange-700'
		};
		return colors[label] || 'bg-gray-100 text-gray-700';
	}

	function getPriorityColor(priority: string): string {
		switch (priority) {
			case 'high': return 'border-l-red-500';
			case 'medium': return 'border-l-yellow-500';
			default: return 'border-l-blue-500';
		}
	}
</script>

<div class="flex gap-4 p-6 overflow-x-auto h-full">
	{#each columns as column}
		<div class="flex-shrink-0 w-72 flex flex-col bg-gray-100 rounded-xl">
			<div class="flex items-center justify-between px-4 py-3">
				<div class="flex items-center gap-2">
					<h3 class="font-semibold text-gray-900 text-sm">{column}</h3>
					<span class="text-xs text-gray-500 bg-gray-200 px-2 py-0.5 rounded-full">
						{getTasksForColumn(column).length}
					</span>
				</div>
			</div>
			<div class="flex-1 overflow-y-auto px-3 pb-3 space-y-2">
				{#each getTasksForColumn(column) as task (task.id)}
					<div
						class="bg-white rounded-lg p-3 border border-gray-200 border-l-4 {getPriorityColor(task.priority)} shadow-sm hover:shadow-md transition-shadow cursor-pointer"
						onclick={() => onEditTask(task)}
					>
						<div class="flex items-start justify-between mb-2">
							<h4 class="font-medium text-gray-900 text-sm flex-1">{task.title}</h4>
							<GripVertical class="w-4 h-4 text-gray-300 flex-shrink-0" />
						</div>
						{#if task.description}
							<p class="text-xs text-gray-500 mb-2 line-clamp-2">{task.description}</p>
						{/if}
						{#if task.labels.length > 0}
							<div class="flex flex-wrap gap-1 mb-2">
								{#each task.labels as label}
									<span class="px-1.5 py-0.5 text-xs font-medium rounded {getLabelColor(label)}">
										{label}
									</span>
								{/each}
							</div>
						{/if}
						<div class="flex items-center justify-between text-xs text-gray-400">
							<span>{task.assignee}</span>
							{#if task.dueDate}
								<div class="flex items-center gap-1">
									<Clock class="w-3 h-3" />
									<span>{task.dueDate}</span>
								</div>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		</div>
	{/each}
</div>
`,
			"src/lib/components/TaskModal.svelte": `<script lang="ts">
	import { X } from 'lucide-svelte';

	interface Task {
		id: string;
		title: string;
		description: string;
		column: string;
		labels: string[];
		assignee: string;
		dueDate: string | null;
		priority: 'low' | 'medium' | 'high';
	}

	interface Props {
		task: Task | null;
		onClose: () => void;
	}

	let { task, onClose }: Props = $props();

	const columns = '{{columns}}'.split(',').map(c => c.trim());
	const availableLabels = '{{labels}}'.split(',').map(l => l.trim());

	let title = $state(task?.title || '');
	let description = $state(task?.description || '');
	let column = $state(task?.column || columns[0]);
	let assignee = $state(task?.assignee || '');
	let dueDate = $state(task?.dueDate || '');
	let priority = $state<'low' | 'medium' | 'high'>(task?.priority || 'medium');

	function handleSubmit(e: Event) {
		e.preventDefault();
		// Save task logic here
		onClose();
	}
</script>

<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onclick={onClose}>
	<div class="bg-white rounded-xl shadow-2xl w-full max-w-lg mx-4" onclick|stopPropagation>
		<div class="flex items-center justify-between p-6 border-b border-gray-200">
			<h2 class="text-lg font-semibold text-gray-900">
				{task ? 'Edit Task' : 'New Task'}
			</h2>
			<button onclick={onClose} class="p-2 text-gray-400 hover:text-gray-600 rounded-lg hover:bg-gray-100">
				<X class="w-5 h-5" />
			</button>
		</div>
		<form onsubmit={handleSubmit} class="p-6 space-y-4">
			<div>
				<label for="title" class="block text-sm font-medium text-gray-700 mb-1">Title</label>
				<input id="title" type="text" bind:value={title} required class="w-full px-4 py-2 border border-gray-300 rounded-lg" />
			</div>
			<div>
				<label for="description" class="block text-sm font-medium text-gray-700 mb-1">Description</label>
				<textarea id="description" bind:value={description} rows="3" class="w-full px-4 py-2 border border-gray-300 rounded-lg"></textarea>
			</div>
			<div class="grid grid-cols-2 gap-4">
				<div>
					<label for="column" class="block text-sm font-medium text-gray-700 mb-1">Column</label>
					<select id="column" bind:value={column} class="w-full px-4 py-2 border border-gray-300 rounded-lg">
						{#each columns as col}
							<option value={col}>{col}</option>
						{/each}
					</select>
				</div>
				<div>
					<label for="priority" class="block text-sm font-medium text-gray-700 mb-1">Priority</label>
					<select id="priority" bind:value={priority} class="w-full px-4 py-2 border border-gray-300 rounded-lg">
						<option value="low">Low</option>
						<option value="medium">Medium</option>
						<option value="high">High</option>
					</select>
				</div>
			</div>
			<div class="grid grid-cols-2 gap-4">
				<div>
					<label for="assignee" class="block text-sm font-medium text-gray-700 mb-1">Assignee</label>
					<input id="assignee" type="text" bind:value={assignee} class="w-full px-4 py-2 border border-gray-300 rounded-lg" />
				</div>
				<div>
					<label for="due-date" class="block text-sm font-medium text-gray-700 mb-1">Due Date</label>
					<input id="due-date" type="date" bind:value={dueDate} class="w-full px-4 py-2 border border-gray-300 rounded-lg" />
				</div>
			</div>
			<div class="flex justify-end gap-3 pt-4">
				<button type="button" onclick={onClose} class="px-4 py-2 text-sm text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200">
					Cancel
				</button>
				<button type="submit" class="px-4 py-2 text-sm text-white rounded-lg" style="background-color: {{primary_color}}">
					{task ? 'Save Changes' : 'Create Task'}
				</button>
			</div>
		</form>
	</div>
</div>
`,
			"package.json": `{
	"name": "{{app_name}}",
	"version": "1.0.0",
	"type": "module",
	"scripts": {
		"dev": "vite dev",
		"build": "vite build",
		"preview": "vite preview"
	},
	"devDependencies": {
		"@sveltejs/adapter-auto": "^3.0.0",
		"@sveltejs/kit": "^2.0.0",
		"svelte": "^5.0.0",
		"tailwindcss": "^3.4.0",
		"typescript": "^5.0.0",
		"vite": "^5.0.0"
	},
	"dependencies": {
		"lucide-svelte": "^0.300.0"
	}
}
`,
		},
	}
}
