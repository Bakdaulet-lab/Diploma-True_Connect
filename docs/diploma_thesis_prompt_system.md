# Diploma Thesis Generation — Master Prompt System
# Halal Dating App (Flutter + Go) — Full Academic Documentation
# ============================================================
#
# STRATEGY: Generate chapter by chapter. DO NOT run all at once.
# Run Prompt 0 first. Then run each numbered prompt separately.
# This prevents output truncation and keeps quality high.
#
# MODE:    Claude Code — Plan mode for Prompt 0, Edit/Auto for 1–7
# MODEL:   claude-opus-4-6 (strongest model, required for 20+ citations)
# ============================================================

═══════════════════════════════════════════════════════════════
PROMPT 0 — CODEBASE RECONNAISSANCE (Run this FIRST, only once)
═══════════════════════════════════════════════════════════════

You are a senior software engineer preparing to write a full academic
diploma thesis for a halal Islamic dating/matrimony mobile application.

YOUR ONLY TASK RIGHT NOW: Read and map the entire codebase.
Do NOT write any documentation yet. Just read and build a knowledge base.

WHY THIS MATTERS: You will write 7 chapters separately. Each chapter
needs accurate technical details from the real code. If you get facts
wrong here, every chapter will be wrong. Take your time.

--- STEP 1: MAP THE FULL STRUCTURE ---

Run these commands and read the output carefully:
  find . -type f -name "*.go" | sort
  find . -type f -name "*.dart" | sort
  find . -name "docker-compose*" -o -name "Dockerfile*" | sort
  find . -name "nginx.conf" -o -name "*.conf" | sort
  find . -name "*.sql" -o -name "*.sh" | sort
  find . -name "*.yml" -o -name "*.yaml" | sort

--- STEP 2: READ EVERY BACKEND GO FILE ---

For each .go file, extract and note:
  - Package name and which domain module it belongs to
  - All exported struct names with their fields and types
  - All exported function/method names with parameter types
  - All SQL queries: exact table names, column names, WHERE conditions
  - All API route registrations (method + path + handler)
  - All error types and constants
  - Any business logic rules (validation, state transitions, checks)

Pay special attention to:
  - The WebSocket hub implementation (how rooms/connections work)
  - The matching/discovery query logic (filters, geolocation)
  - The auth middleware (what it validates, what it injects into context)
  - The contract between modules (interfaces / client types)

--- STEP 3: READ EVERY FRONTEND DART FILE ---

For each .dart file, extract and note:
  - Screen names and their navigation routes
  - Provider names and the state they manage
  - Model class names with all fields
  - API endpoint constants (exact URLs)
  - WebSocket message types sent and received
  - How auth tokens are stored and retrieved
  - State management patterns used

--- STEP 4: READ ALL INFRASTRUCTURE FILES ---

Read in full: docker-compose.yml, Dockerfile, nginx.conf,
any GitHub Actions .yml files, any database migration files.

Note: service names, port mappings, volume mounts, environment vars,
nginx proxy rules, WebSocket upgrade headers, rate limiting rules.

--- STEP 5: BUILD THE KNOWLEDGE BASE FILE ---

Create the file: docs/codebase_map.md

Write it in this exact structure:

# Codebase Knowledge Base — Halal Dating App

## Project Identity
- App name: [from code]
- Backend language/framework: Go + Gin (or whatever you find)
- Frontend: Flutter
- Database: PostgreSQL + Redis (confirm from docker-compose)
- Authentication: JWT (confirm details: algorithm, expiry times)

## Backend Architecture
- Architecture pattern: [Modular Monolith / Clean Architecture / DDD]
- Layer structure: [Handler → Service → Repository → Model, or what you find]
- Inter-module communication: [interfaces / direct calls / events]

## Backend Modules
[For each module/package found — write a block like this:]

### [module name] Module
- Handler: [file path]
- Service: [file path]
- Repository: [file path]
- Model: [file path]
- Key structs:
  - StructName: field1 type, field2 type, ...
- Key service functions:
  - FunctionName(params) returns
- API endpoints:
  - METHOD /path/to/endpoint → HandlerFunction
- DB tables used: [table names]
- Business rules: [any domain logic found]

## Database Schema
[For every table found in migrations or models:]

### table_name
| Column | Type | Constraints |
|--------|------|-------------|
| id | uuid/serial | PRIMARY KEY |
| ... | ... | ... |

Foreign Keys: [list FK relationships]

## API Endpoints — Complete List
[Every single endpoint:]
METHOD /path — description — Auth: yes/no — Role: customer/supplier/any

## WebSocket Protocol
- Connection URL: [from constants]
- Auth mechanism: [how the first message works]
- Message types (incoming): [list with JSON structure]
- Message types (outgoing): [list with JSON structure]
- Hub structure: [how connections are tracked]
- Redis pub/sub: [channel naming, how messages are routed]

## Frontend Screens
[List every screen file with: name, route, purpose, provider used]

## Frontend Providers (State Management)
[List every provider: name, state type, key methods]

## Key Model Classes (Frontend)
[List every model class with fields]

## Infrastructure
- Docker services: [names, images, ports]
- nginx: [proxy rules, WebSocket handling, rate limits]
- CI/CD: [pipeline stages if found]
- Environment variables: [list keys, not values]

## Islamic Domain Features Found
[List every feature specific to the Islamic/halal domain:]
- Niyyah (intention): [how it's stored, filtered, used]
- Madhab (school): [how it's stored, filtered, used]
- Mahram chat: [technical implementation]
- Modesty levels: [how it works]
- [Any other Islamic-specific feature]

## Known Code Issues / Technical Debt
[Any bugs, incomplete implementations, or TODOs you notice]

---

After saving docs/codebase_map.md, report:
"Reconnaissance complete. Knowledge base saved to docs/codebase_map.md.
Found [N] backend modules, [N] API endpoints, [N] Flutter screens.
Ready to write Chapter 1."

DO NOT start writing documentation. Wait for the next prompt.


═══════════════════════════════════════════════════════════════
PROMPT 1 — Chapter 1: Introduction
(Run after Prompt 0 is complete)
═══════════════════════════════════════════════════════════════

You are writing Chapter 1 of a diploma thesis for a halal Islamic
dating/matrimony mobile application built with Flutter and Go.

BEFORE WRITING:
1. Read docs/codebase_map.md completely
2. Re-read the main app entry file (main.dart or app.dart)
3. Re-read the router/navigation file
4. Re-read the constants file (API URLs, app name)

GRADING: Chapter 1 is worth 15 points.
The grader expects: project overview, significance/impact, objectives.
Generic filler = zero. Specific + technical = full marks.

CONSTRAINTS:
- Minimum 1,500 words (count before saving)
- Formal academic English throughout
- Reference actual features from docs/codebase_map.md
- No LaTeX syntax — plain Markdown only
- No web frontend — this is a mobile app only

WRITE the following (do not skip any section):

─────────────────────────────────────────────────────────────
# Chapter 1: Introduction
─────────────────────────────────────────────────────────────

## 1.1 Background and Motivation

Write 300+ words covering:
- The global Muslim population (1.8+ billion) and the challenge
  of finding a spouse that meets Islamic requirements in the
  modern world
- How mainstream dating apps (Tinder, Bumble) fundamentally
  violate Islamic principles: free mixing of genders, photo-first
  culture, casual intent, no guardian/mahram involvement
- How "Islamic" apps on the market (Muzz, Salams, Hawaya) still
  fall short: no real mahram chat supervision, no madhab preference
  filtering, no niyyah (intention) verification, often just
  rebranded Western dating apps
- The specific cultural context of Kazakhstan and Central Asia:
  Muslim-majority population, growing smartphone penetration,
  lack of a locally-built solution
- The technical argument: a platform built with Islamic values
  as first-class architectural requirements, not as UI cosmetics

## 1.2 Problem Statement

Write 250+ words. State exactly 5 precise problems this project solves.
Format each as: "Problem N: [specific problem statement]"
Derive these problems from the actual features you found in the code.
Examples (verify each exists before including):
- The absence of mahram (guardian) supervision in existing chat systems
- The inability to filter potential matches by madhab (Islamic school of thought)
- No declaration of niyyah (serious marriage intent vs casual)
- Mainstream apps incentivize superficial swiping over meaningful matching
- No real-time communication infrastructure designed for Islamic privacy norms
End with a concise one-paragraph formal problem statement.

## 1.3 Significance and Impact

Write 250+ words on:
- Social impact: what it means for Muslim users to have a platform
  that respects Islamic values — specific scenarios
- Technical significance: what this project contributes to the
  field of socially-conscious app design
- Economic opportunity: the underserved market in Kazakhstan,
  the potential for regional expansion
- Research contribution: what can be learned from building
  Clean Architecture + DDD around an Islamic domain model

## 1.4 Research Objectives

Write 200+ words. List exactly 7-8 objectives.
Each must follow this format:
"To [verb] [specific thing] using [specific technology]
 in order to [measurable outcome]"

Base each objective on a real feature found in docs/codebase_map.md.
Do not invent objectives for features that don't exist.

## 1.5 Scope and Limitations

Write 200+ words.

In scope: [list every major feature confirmed in the codebase]
Out of scope: [list what was explicitly NOT built — web client,
  payments, video calling — confirm these are absent from the code]
Current limitations: [list real technical limitations you observed
  while reading the code — be honest, this shows intellectual integrity]

## 1.6 Structure of the Document

Write 100+ words.
One paragraph per chapter (Chapters 1-7) describing what it covers.
─────────────────────────────────────────────────────────────

SAVE to: docs/chapter1.md
VERIFY word count is ≥ 1,500 before saving.
REPORT: "Chapter 1 complete. Word count: [N]. Saved to docs/chapter1.md"


═══════════════════════════════════════════════════════════════
PROMPT 2 — Chapter 2: Literature Review
(Run after Chapter 1 is confirmed complete)
═══════════════════════════════════════════════════════════════

You are writing Chapter 2 of a diploma thesis for a halal Islamic
dating/matrimony mobile application.

BEFORE WRITING:
1. Read docs/codebase_map.md

GRADING: This chapter is worth 15 points.
THE SINGLE MOST IMPORTANT REQUIREMENT: minimum 20 academic citations.
The grader will count them. Missing this = automatic point deduction.
Count your citations yourself before saving.

CITATION FORMAT:
- Inline: [N] (IEEE numbered style)
- Full reference list → append to docs/references.md
- Full reference format: [N] A. Author, "Title," Journal/Publisher, year.

CONSTRAINTS:
- Minimum 3,000 words
- 20+ citations (you must count and verify before saving)
- Each section must synthesize the literature — not just list papers
- Write genuine analysis connecting each paper to this project
- No LaTeX — plain Markdown only

─────────────────────────────────────────────────────────────
# Chapter 2: Literature Review
─────────────────────────────────────────────────────────────

## 2.1 Online Dating Platforms: Evolution and Behavioral Research
[Target: ~550 words | Minimum citations: 5]

Synthesize research on:
- The transition from traditional to online matchmaking (historical)
- User motivations and self-presentation in online dating profiles
- Algorithmic matching: how platforms decide who sees whom
- Trust, safety, and deception in online dating
- The gamification of dating (swipe mechanics and their behavioral effects)

Cite these real papers (use accurate details):
- E.J. Finkel, P.W. Eastwick, B.R. Karney, H.T. Reis, S. Sprecher,
  "Online Dating: A Critical Analysis from the Perspective of
  Psychological Science," Psychological Science in the Public Interest,
  vol. 13, no. 1, pp. 3-66, 2012.
- G. Tyson, V.C. Perta, H. Haddadi, M.C. Seto, "A First Look at
  User Activity on Tinder," IEEE/ACM International Conference on
  Advances in Social Networks Analysis and Mining, 2016.
- G.J. Hitsch, A. Hortaçsu, D. Ariely, "Matching and Sorting in
  Online Dating," American Economic Review, vol. 100, no. 1,
  pp. 130-163, 2010.
- Add 2+ more real papers on online dating behavior or algorithms

Connect each paper to your project: "Unlike [paper], this system
prioritizes X because of Islamic requirement Y."

## 2.2 Islamic Marriage Practices and the Role of Technology
[Target: ~500 words | Minimum citations: 4]

Synthesize research on:
- The Islamic framework for marriage: role of wali (guardian),
  mahr (dowry), niyyah (intention), concept of mahram
- Academic research on Muslim communities and digital technology use
- Studies on halal certification concepts applied to digital services
- Cultural considerations in designing apps for Muslim-majority markets
- Research on digital Islam and online religious practice

Find and cite real academic sources. If specific papers are uncertain,
cite broader works:
- Books or papers on Islam and modernity
- Pew Research Center reports on Muslim population and technology use
- Academic papers on Muslim digital communities
- Papers on culturally-sensitive app design

## 2.3 Recommendation and Matching Algorithms
[Target: ~500 words | Minimum citations: 4]

Synthesize:
- Collaborative filtering and its limitations (cold-start problem)
- Content-based filtering using user attributes
- Hybrid recommender systems
- Location-based social networks and geolocation recommendations
- Matching market theory (Gale-Shapley stable matching)
- Preference-based filtering in social platforms

Cite real papers:
- J. Bobadilla, F. Ortega, A. Hernando, A. Gutiérrez, "Recommender
  Systems Survey," Knowledge-Based Systems, vol. 46, pp. 109-132, 2013.
- D. Gale, L.S. Shapley, "College Admissions and the Stability
  of Marriage," American Mathematical Monthly, vol. 69, no. 1,
  pp. 9-15, 1962.
- Papers on geolocation-based recommendation (cite LBSN research)
- Papers on attribute-based filtering in social matching

Connect to project: explain how niyyah/madhab filters implement
a form of constraint-based preference filtering.

## 2.4 Real-Time Communication Architecture
[Target: ~450 words | Minimum citations: 3]

Synthesize:
- Evolution of real-time web: polling → long-polling → SSE → WebSocket
- WebSocket protocol specification and use cases
- Redis pub/sub as a message broker for horizontal scaling
- Challenges: message ordering, delivery guarantees, reconnection
- Mobile-specific considerations: battery, network switching, background

Cite:
- I. Fette, A. Melnikov, "The WebSocket Protocol," IETF RFC 6455, 2011.
- Research on Redis architecture and pub/sub performance
- Papers on scalable real-time messaging systems
- Any paper on mobile chat architecture or instant messaging at scale

Explain specifically how the Hub pattern + Redis enables your
system to route chat messages between users.

## 2.5 Software Architecture for Complex Domains
[Target: ~500 words | Minimum citations: 4]

Synthesize:
- Clean Architecture and separation of concerns
- Domain-Driven Design: bounded contexts, aggregates, ubiquitous language
- Modular monolith architecture: benefits over both monolith and microservices
- Dependency inversion and testability
- When to use microservices vs modular monolith (scale thresholds)

Cite:
- R.C. Martin, "Clean Architecture: A Craftsman's Guide to Software
  Structure and Design," Prentice Hall, 2017.
- E. Evans, "Domain-Driven Design: Tackling Complexity in the Heart
  of Software," Addison-Wesley, 2003.
- S. Newman, "Building Microservices," 2nd ed., O'Reilly, 2021.
- M. Fowler, J. Lewis, "Microservices," martinfowler.com, 2014.

Show specifically how DDD maps to the halal domain:
bounded contexts = auth/discovery/chat/feed/mahram etc.,
ubiquitous language = niyyah/mahram/madhab in the code.

## 2.6 Cross-Platform Mobile Development
[Target: ~350 words | Minimum citations: 3]

Synthesize:
- History: Cordova → React Native → Flutter → KMM
- Flutter's architectural advantage: compiled to native ARM,
  own rendering engine (Skia/Impeller), no JavaScript bridge
- State management evolution in Flutter: Provider → BLoC → Riverpod
- Performance comparison: Flutter vs React Native on animation-heavy UIs

Cite:
- Google, "Flutter: Beautiful Native Apps in Record Time,"
  flutter.dev, 2018+ (technical documentation)
- A. Biørn-Hansen, T.A. Majchrzak, T.-M. Grønli, "Progressive Web
  Apps vs. Native Mobile Apps: A Multi-criteria Comparison,"
  MobiWis Conference, 2017.
- Any peer-reviewed paper comparing Flutter with React Native
  on performance metrics

Justify why Flutter was the right choice for this specific
type of app (animation-heavy swipe interface, complex state).

## 2.7 Literature Review Summary
[Target: ~200 words]

Synthesize what the literature tells us:
- What is known about online matchmaking
- What is known about Islamic digital practice
- What architectural patterns are proven for social platforms
- The specific gap that this project addresses that the literature
  has not yet solved

─────────────────────────────────────────────────────────────

BEFORE SAVING — CHECK:
1. Count total citations: should be ≥ 20
2. Count words: should be ≥ 3,000
3. Verify every [N] has a matching entry in the references list

SAVE chapter to: docs/chapter2.md
APPEND all references to: docs/references.md (number from where you left off)
REPORT: "Chapter 2 complete. Citations: [N]. Words: [N]."


═══════════════════════════════════════════════════════════════
PROMPT 3 — Chapter 3: Analysis of Existing Systems
(Run after Chapter 2 is confirmed complete)
═══════════════════════════════════════════════════════════════

You are writing Chapter 3 of a diploma thesis.

BEFORE WRITING:
1. Read docs/codebase_map.md
2. Focus on the Islamic-specific features listed there

GRADING: 10 points. Grader expects: SWOT analysis, gap analysis,
comparison table, data collection methods, challenges.

CONSTRAINTS:
- Minimum 2,500 words
- Include at least one formal comparison table
- Include a formal SWOT analysis (formatted as a 2x2 table)
- No LaTeX — plain Markdown

─────────────────────────────────────────────────────────────
# Chapter 3: Analysis of Existing Systems
─────────────────────────────────────────────────────────────

## 3.1 Overview of the Competitive Landscape

~200 words. Categorize competitors:
- Mainstream dating apps (not designed for Muslims)
- Islamic matrimony apps (designed for Muslims but with limitations)
- Regional matchmaking platforms (Kazakhstan/Central Asia)
Explain why analysis of all three categories is necessary.

## 3.2 Analysis of Individual Systems

For each platform, write 200-250 words covering:
features, target audience, Islamic compliance level,
geographic reach, known technical approach, key weaknesses.

### 3.2.1 Tinder
### 3.2.2 Bumble
### 3.2.3 Muzz (formerly Muzmatch)
### 3.2.4 Salams (formerly Minder)
### 3.2.5 Hawaya (by Match Group)
### 3.2.6 Regional/Traditional Matchmaking Platforms

## 3.3 Functional Comparison Table

Create a comprehensive Markdown table:

| Feature | Tinder | Bumble | Muzz | Salams | Hawaya | This Project |
|---------|--------|--------|------|--------|--------|--------------|
| Mahram/guardian chat supervision | No | No | Limited | No | No | Yes |
| Niyyah (intention) declaration | No | No | Partial | No | No | Yes |
| Madhab school filtering | No | No | No | No | No | Yes |
| Community feed / Islamic content | No | No | No | No | No | Yes |
| Photo privacy / modesty controls | No | No | Partial | Partial | Partial | Yes |
| Real-time WebSocket chat | Yes | Yes | Yes | Yes | Yes | Yes |
| Distance-based discovery | Yes | Yes | Yes | Yes | Yes | Yes |
| Kazakhstan/Central Asia focus | No | No | No | No | No | Yes |
| Bilingual (Kazakh/Russian) | No | No | No | No | No | [check code] |
| Open source / local hosting | No | No | No | No | No | Yes |

[Add more rows based on features found in the codebase]

After the table, write 200 words analyzing what the table reveals.

## 3.4 SWOT Analysis

Present as a formatted 2x2 table, then expand each quadrant
with 100-150 words of explanation:

| | Helpful | Harmful |
|---|---|---|
| Internal | **Strengths** | **Weaknesses** |
| External | **Opportunities** | **Threats** |

Strengths (derive from actual codebase):
- [List specific technical strengths found in code]
- Mahram chat implementation
- Clean Architecture enabling testability
- Domain-specific filters (niyyah, madhab)
- Open/self-hosted: no vendor lock-in

Weaknesses (be honest about what you found):
- [Real limitations from code]
- [Missing features observed]
- [Technical debt noticed]

Opportunities:
- 1.8B global Muslim market
- Kazakhstan's growing smartphone penetration
- No localized Islamic matrimony app in the region
- Government digitalization initiatives

Threats:
- Muzz's established user base and funding
- Cultural resistance to digital matchmaking
- App store policies on dating apps
- User privacy concerns

## 3.5 Gap Analysis

~400 words. Create a formal gap analysis table:

| Gap in Existing Systems | How This Project Addresses It | Feature in Code |
|------------------------|------------------------------|-----------------|
| No mahram supervision in chat | Mahram chat module | [file name] |
| No madhab filtering | Madhab filter in discovery | [file name] |
| No niyyah declaration | Niyyah profile field + filter | [file name] |
| [add more from codebase] | ... | ... |

After the table, write a narrative connecting the gaps to the
design decisions visible in the codebase.

## 3.6 Data Collection and User Research Methodology

~300 words. Describe:
- What user research was conducted (survey, interviews, or planned)
- Target respondents: Muslim users in Kazakhstan seeking marriage
- Key findings that shaped the feature set
- How research findings map to implemented features
(If no formal research was done, describe what would be needed
and how the feature set reflects implicit user needs.)

## 3.7 Challenges Unique to Islamic Matrimony Applications

~300 words covering:
- Privacy challenges: photo visibility (mahram must approve
  before non-mahram can see full photo?)
- Trust challenges: how to verify Islamic identity/practice
- Cultural challenges: balancing tradition with digital UX
- Technical challenges: implementing guardian supervision
  without compromising real-time performance
- Regulatory challenges: app store policies, data privacy laws

─────────────────────────────────────────────────────────────
SAVE to: docs/chapter3.md
REPORT: word count when done.


═══════════════════════════════════════════════════════════════
PROMPT 4 — Chapter 4: Methodology
(Run after Chapter 3 is confirmed complete)
═══════════════════════════════════════════════════════════════

You are writing Chapter 4 of a diploma thesis.

BEFORE WRITING:
1. Read docs/codebase_map.md
2. Re-read the module structure section carefully

GRADING: 15 points. Grader expects: development methodology
(Agile/Scrum), DDD explanation, data analysis, justification.

CONSTRAINTS:
- Minimum 2,000 words
- Must explain DDD with specific examples from the actual codebase
- Include a requirements table (functional + non-functional)
- No LaTeX — plain Markdown

─────────────────────────────────────────────────────────────
# Chapter 4: Methodology
─────────────────────────────────────────────────────────────

## 4.1 Development Methodology: Agile with Scrum

~400 words covering:
- Why Agile was appropriate for this project (uncertain requirements
  in a novel domain, need for iterative user feedback)
- Sprint structure: how work was organized into iterations
- User stories organized around the Islamic domain:
  "As a Muslim user seeking marriage, I want to declare my niyyah
   so that I only see matches with serious marriage intent"
- Backlog prioritization: which MVP features came first
  (auth → profile → discovery → matching → chat → feed → mahram)
- How sprints enabled progressive refinement of Islamic-specific features
- Retrospective learning examples

## 4.2 Domain-Driven Design Application

~600 words — the most important section of this chapter.

4.2.1 Identifying Bounded Contexts
Map each backend module to a DDD bounded context:
(Use exact module names from docs/codebase_map.md)
- Auth Context: [describe]
- User/Profile Context: [describe]
- Discovery Context: [describe what the matching/discovery module does]
- Feed Context: [describe]
- Chat Context: [describe]
- Mahram Context: [describe]
- Settings Context: [describe]

Explain how context boundaries were enforced in the codebase
(package-level isolation, interface-based communication).

4.2.2 Ubiquitous Language
List the domain terms used consistently throughout the codebase:
- niyyah: [definition + where it appears in code]
- mahram: [definition + where it appears in code]
- madhab: [definition + where it appears in code]
- match: [definition + where it appears in code]
- like/dislike: [definition + where it appears in code]
Show how these terms appear identically in Go struct names,
database columns, API endpoints, and Flutter model fields.
This is the essence of DDD's ubiquitous language.

4.2.3 Aggregates and Entities
Identify the key aggregates from the domain model:
- User aggregate: [root entity + what it owns]
- Match aggregate: [how a match is formed from two likes]
- Conversation aggregate: [messages within a match]
- Post aggregate: [feed post + comments + likes]
Explain the consistency boundaries each aggregate enforces.

4.2.4 Repository Pattern
Describe how the repository pattern isolates domain logic
from database concerns. Give specific examples from the code
(e.g., IDiscoveryRepository, IChatRepository, etc.)

## 4.3 Requirements Engineering

~400 words.

Create a Functional Requirements table:
| ID | Requirement | Priority | Source |
|----|-------------|----------|--------|
| FR-01 | System shall allow users to register with phone/email | High | Auth module |
| FR-02 | System shall support niyyah declaration during onboarding | High | Onboarding |
| FR-03 | System shall filter discovery results by madhab | High | Discovery |
| [add all major functional requirements from codebase] | | | |

Create a Non-Functional Requirements table:
| ID | Requirement | Metric | Implementation |
|----|-------------|--------|----------------|
| NFR-01 | Real-time message delivery | < 200ms latency | WebSocket + Redis |
| NFR-02 | Authentication security | JWT + bcrypt | Auth module |
| NFR-03 | API response time | < 500ms p95 | Go + PostgreSQL |
| [add more] | | | |

## 4.4 Testing Strategy

~300 words covering:
- Unit testing approach: which service layer functions were tested
- Integration testing: API endpoint testing approach
- Manual testing: how WebSocket flows were validated
- Test coverage goals vs actual coverage
- Edge cases specifically tested for Islamic domain logic
  (e.g., what happens when mahram is removed mid-conversation)

## 4.5 Justification of Methodology

~300 words.
Argue why Agile + DDD was the correct methodology for THIS specific project:
- Agile handles uncertainty in novel domain (no existing template
  for halal matrimony apps with mahram chat)
- DDD ensures the codebase reflects Islamic domain concepts faithfully,
  not as an afterthought
- The combination of Agile iteration + DDD modeling enabled
  progressive discovery of domain rules
- Compare briefly to alternatives: Waterfall (too rigid for novel domain),
  pure technical decomposition (would lose domain meaning)

─────────────────────────────────────────────────────────────
SAVE to: docs/chapter4.md
REPORT: word count when done.


═══════════════════════════════════════════════════════════════
PROMPT 5 — Chapter 5: Architecture and UML
(Run after Chapter 4 is confirmed complete)
═══════════════════════════════════════════════════════════════

You are writing Chapter 5 of a diploma thesis.

BEFORE WRITING:
1. Read docs/codebase_map.md completely
2. Re-read: the hub/websocket file, the main router file,
   all service interface files, all model files,
   docker-compose.yml, nginx config

GRADING: 15 points. CRUCIAL CHAPTER.
Grader expects: system architecture detail, UML diagram specifications,
database design, real technical accuracy.

CONSTRAINTS:
- Minimum 3,000 words
- Must describe all UML diagrams in enough detail for a developer
  to draw them without asking questions
- All struct/function/table names must match the real codebase
- No LaTeX — plain Markdown

─────────────────────────────────────────────────────────────
# Chapter 5: System Architecture and Design
─────────────────────────────────────────────────────────────

## 5.1 System Architecture Overview

~300 words + architecture description.

Describe the three-tier architecture:
1. Flutter Mobile Client (presentation + state management)
2. Go Backend (modular monolith with clean architecture layers)
3. Data layer (PostgreSQL for persistence, Redis for pub/sub/cache)

Explain the communication channels:
- HTTPS REST API: Flutter ↔ Go backend (all CRUD operations)
- WebSocket WSS: Flutter ↔ Go Hub (real-time chat only)
- Internal: Go Hub ↔ Redis pub/sub (cross-connection routing)

Explain why Modular Monolith was chosen over microservices at this stage.

## 5.2 Backend Clean Architecture Layers

~400 words. For each layer, explain its responsibility and give
real examples from the codebase:

Handler Layer (API / Presentation):
- Receives HTTP requests, validates input, calls service
- Returns JSON responses with consistent structure
- Applies middleware (JWT auth, rate limiting, logging)
- Example: [HandlerFunctionName from code]

Service Layer (Business Logic):
- Orchestrates domain use cases
- Enforces business rules and invariants
- Coordinates between repositories and other services
- Example: [ServiceFunctionName from code] with explanation of its logic

Repository Layer (Data Access):
- Abstracts all database operations
- Implements interfaces defined by the service layer
- Translates domain models to SQL queries
- Example: [RepositoryFunctionName from code]

Model Layer (Domain Objects):
- Domain entities and value objects
- DTOs for API transport
- Validation rules
- Example: [ModelStructName from code] with fields

Client/Interface Layer (Inter-module):
- Interfaces that define contracts between modules
- Prevents tight coupling between bounded contexts
- Example: [InterfaceName from code]

## 5.3 Backend Module Architecture

~600 words. For each major module, write a paragraph describing:
- Domain responsibility
- Key service functions and their logic
- Repository functions and queries
- How it interacts with other modules

Cover ALL modules found in docs/codebase_map.md.
Reference actual function and struct names.

## 5.4 Database Architecture

~400 words.

List every table with description of its purpose.
Then provide key design decision explanations:
- Why UUIDs vs serial integers for primary keys
- How the like/match relationship is modeled
  (what prevents duplicate likes, how mutual detection works)
- How messages are stored (table structure, indexes for performance)
- How geolocation data is stored and queried
- Index strategy: which columns are indexed and why

## 5.5 WebSocket Architecture

~300 words.

Describe the Hub pattern implementation:
- How client connections are registered and unregistered
- How the hub routes incoming messages to the correct recipient
- How Redis pub/sub extends routing across potential server instances
- The authentication handshake: first message must be {type: "auth"}
- Connection lifecycle: connect → auth → subscribe → receive → disconnect
- Message envelope structure: type, payload, timestamp

## 5.6 UML Diagram Specifications

This section provides complete textual specifications for all UML
diagrams. Each specification must be detailed enough to draw
without asking any questions.

### Diagram A: Use Case Diagram

Actors:
1. Muslim User (primary — seeks marriage)
2. Mahram Guardian (secondary — supervises)
3. System Administrator (secondary — manages platform)

Use Cases (list every one, with <<include>> and <<extend>>):
[Generate based on features found in codebase]

Example relationships to include:
- "View Discovery Feed" <<include>> "Authenticate"
- "Send Message" <<include>> "Verify Match Exists"
- "Send Mahram Message" <<extend>> "Send Message"
- etc.

### Diagram B: System Architecture Diagram

Nodes to include:
- Flutter Mobile App (left side)
  - Presentation Layer (screens)
  - State Layer (Riverpod providers)
  - Service Layer (API services)
  - Model Layer
- Go Backend (center)
  - nginx (reverse proxy)
  - Each module box: Auth, Discovery, Chat, Feed, Mahram, Settings, etc.
- PostgreSQL (right, bottom)
- Redis (right, middle)

Connections with labels:
- Flutter → nginx: HTTPS/REST + WSS
- nginx → Go modules: HTTP proxy
- Go Chat Hub → Redis: pub/sub PUBLISH/SUBSCRIBE
- Go modules → PostgreSQL: SQL queries
- Arrows for: JWT token flow, WebSocket upgrade

### Diagram C: Sequence Diagram — User Registration and Onboarding

Participants: User, Flutter App, Go Backend (Auth Module),
              Go Backend (User Module), PostgreSQL

Step-by-step sequence:
1. User enters phone/email + password
2. Flutter → POST /auth/register → Auth Handler
3. Auth Service validates input
4. Auth Repository → INSERT users table → PostgreSQL
5. JWT access token generated
6. Refresh token set as HttpOnly cookie
7. Response: {access_token, user_id}
8. Flutter stores access_token in secure storage
9. Flutter navigates to onboarding screens
10. User selects niyyah, madhab, modesty level
11. Flutter → PATCH /users/profile → User Handler
12. User Repository → UPDATE profiles → PostgreSQL
13. Response: {updated_user}
14. Flutter → navigate to discovery

[Add exact function names from codebase to each step]

### Diagram D: Sequence Diagram — Swipe Right and Match Creation

Participants: User A, Flutter App, Go Backend (Discovery Module),
              Go Backend (Match Module), PostgreSQL, WebSocket Hub

Step-by-step sequence:
[Derive from actual like/match logic in codebase]
Include: the mutual check, match creation, WebSocket notification to both users

### Diagram E: Sequence Diagram — Real-time Chat Message

Participants: User A (sender), Flutter A, WebSocket Hub,
              Redis, WebSocket Hub Instance B, Flutter B, User B (receiver)

Step-by-step sequence:
1. User A types message, taps Send
2. Flutter A → WebSocket sink.add({type: "chat_msg", payload: {...}})
3. Hub receives message, looks up recipient connection
4. If recipient connected locally: direct delivery
5. If recipient on different server: Hub → Redis PUBLISH channel
6. Redis → SUBSCRIBE → Hub B receives message
7. Hub B → WebSocket → Flutter B
8. Flutter B → UI update (new message appears)
9. Hub A → PostgreSQL: INSERT messages table (persist)
10. Hub A → Flutter A: delivery confirmation (or optimistic only)

[Adjust based on actual implementation in the Hub code]

### Diagram F: Sequence Diagram — JWT Token Refresh

Participants: Flutter App, Dio Interceptor, Go Backend (Auth Module),
              PostgreSQL, Secure Storage

Step-by-step sequence:
[Derive from the interceptor code and auth refresh handler]

### Diagram G: Entity-Relationship Diagram (ERD)

For each table in the database, specify:
- Table name
- All columns with data types and constraints
- Primary keys
- Foreign keys with referenced table and column
- Cardinality: one-to-many, many-to-many relationships

[Get exact table names and columns from docs/codebase_map.md]

Relationships to show:
- users 1 ── ∞ likes (a user can like many people)
- users ∞ ── ∞ users (many-to-many via matches table)
- matches 1 ── ∞ messages
- users 1 ── ∞ posts
- posts 1 ── ∞ comments
- [add all relationships from schema]

### Diagram H: Class Diagram (Backend Service Interfaces)

Show the key service interfaces and their implementations:

Interface IAuthService:
- Login(ctx, phone, password) → (token, error)
- Register(ctx, user) → (token, error)
- RefreshToken(ctx, refreshToken) → (token, error)

Interface IDiscoveryService:
- GetCandidates(ctx, userID, filters) → ([]Profile, error)
- RecordLike(ctx, likerID, likedID) → (match?, error)

Interface IChatHub:
- RegisterClient(conn, userID)
- UnregisterClient(userID)
- RouteMessage(msg) → error

[Add all interfaces found in codebase with their methods]

Show dependencies between services with arrows.

─────────────────────────────────────────────────────────────
SAVE to: docs/chapter5.md
REPORT: word count when done.


═══════════════════════════════════════════════════════════════
PROMPT 6 — Chapter 6: Technology Comparison
(Run after Chapter 5 is confirmed complete)
═══════════════════════════════════════════════════════════════

You are writing Chapter 6 of a diploma thesis.

BEFORE WRITING: Read docs/codebase_map.md

GRADING: 10 points. Grader expects: Go vs alternatives,
Flutter vs alternatives, PostgreSQL vs alternatives, justified choices.

CONSTRAINTS:
- Minimum 2,000 words
- Include comparison tables for each technology choice
- Cite real benchmarks or papers where possible
- Every conclusion must be justified by a specific technical argument

─────────────────────────────────────────────────────────────
# Chapter 6: Technology Selection and Justification
─────────────────────────────────────────────────────────────

## 6.1 Backend Language: Go vs Node.js vs Python

Comparison table:
| Criterion | Go | Node.js | Python (FastAPI) |
|-----------|-----|---------|-----------------|
| Concurrency model | Goroutines (M:N threading) | Event loop (single-threaded) | Async/await + GIL |
| Throughput (req/s) | ~150,000 | ~80,000 | ~40,000 |
| Memory footprint | Very low (~10MB idle) | Moderate (~50MB) | High (~80MB) |
| Type safety | Strong (compiled) | Optional (TypeScript) | Optional (hints) |
| Compile-time errors | Yes | No | No |
| WebSocket support | Native goroutines | libuv callbacks | Asyncio |
| Cold start time | Fast | Moderate | Slow |
| Team expertise | [assumed from project] | Common | Common |

~400 words narrative justifying Go specifically for:
- The combination of REST API + WebSocket Hub + Redis pub/sub
  in a single binary (goroutines handle all three naturally)
- Memory efficiency matters at scale on a startup budget
- Compiled binary simplifies Docker image and deployment
- Strong typing catches domain logic errors at compile time
  (critical for Islamic business rules)

## 6.2 Mobile Framework: Flutter vs React Native vs Native

Comparison table:
| Criterion | Flutter | React Native | Native (Kotlin/Swift) |
|-----------|---------|--------------|----------------------|
| Rendering | Own engine (Skia/Impeller) | JavaScript bridge | Platform native |
| Performance | Near-native | Slightly slower | Native |
| Code sharing | ~95% shared | ~80% shared | 0% shared |
| Animation quality | Excellent | Good | Excellent |
| Hot reload | Yes | Yes | No |
| State management | Riverpod/BLoC | Redux/MobX | Platform-specific |
| Community size | Large + growing | Large | Split |
| Dart learning curve | Moderate | Low (JS developers) | Platform-specific |

~400 words justifying Flutter specifically for:
- Swipe-card UI in dating apps requires 60fps animations:
  Flutter's own rendering engine vs JS bridge overhead
- Riverpod provides compile-safe dependency injection
  (important for the many providers in this app)
- Single codebase: future iOS release with no extra cost
- Hot reload accelerated development of complex UI flows

## 6.3 Database: PostgreSQL vs MongoDB vs Firebase Firestore

Comparison table:
| Criterion | PostgreSQL | MongoDB | Firebase Firestore |
|-----------|-----------|---------|-------------------|
| Schema | Strict (migrations) | Flexible (schemaless) | Flexible |
| ACID compliance | Full | Partial (4.0+) | Limited |
| Geolocation queries | PostGIS extension | Native geospatial | Limited |
| JOIN performance | Excellent | Poor (no real JOINs) | No JOINs |
| Relational integrity | Foreign keys enforced | Application-level | Application-level |
| Hosting | Self-hosted | Self-hosted / Atlas | Google Cloud only |
| Cost at scale | Predictable | Variable | Variable + vendor lock |

~400 words justifying PostgreSQL for:
- The data is highly relational: users → likes → matches → messages
  This structure requires JOINs and foreign key integrity
- Geolocation: distance queries for discovery need
  proper spatial indexing (Haversine formula in SQL or PostGIS)
- Self-hosted: control over data (important for privacy-sensitive
  Islamic matrimony data — users don't want their data on
  a foreign cloud service)
- ACID guarantees prevent inconsistent match states
  (two users simultaneously liking each other must create
   exactly one match, not zero or two)

## 6.4 Social Graph: Relational vs Neo4j

~400 words covering:
- The social graph problem in this app:
  users → likes → matches → conversations → messages → mahram access
- When Neo4j wins: graph traversal queries (friends-of-friends,
  6-degrees of separation, complex graph algorithms)
- When PostgreSQL wins: this use case (simple 1-2 hop relationships,
  primary access pattern is "who did user X like" not deep traversal)
- Performance comparison: for 2-hop queries (find mutual connections),
  PostgreSQL with proper indexes is competitive with Neo4j
- Decision: PostgreSQL is sufficient for the current scale and
  use patterns; Neo4j would be considered at 10M+ users with
  complex recommendation graph traversal requirements

## 6.5 Real-Time Communication: WebSocket vs Alternatives

Comparison table:
| Approach | Latency | Server load | Bi-directional | Mobile battery |
|----------|---------|-------------|----------------|----------------|
| Short polling | High (1-5s) | Very high | No | Poor |
| Long polling | Medium | High | No | Moderate |
| Server-Sent Events | Low | Low | No (server→client only) | Good |
| WebSocket | Very low | Low | Yes | Good |
| Firebase RTDB | Very low | Low (Google managed) | Yes | Good |

~250 words justifying WebSocket + Redis pub/sub over Firebase:
- WebSocket gives full control over message routing logic
  (needed for mahram chat: complex access control rules)
- Redis pub/sub enables horizontal scaling without vendor lock-in
- Firebase RTDB would mean user messages stored on Google servers:
  unacceptable for privacy-sensitive matrimony conversations
- The Hub pattern cleanly separates connection management
  from message routing

─────────────────────────────────────────────────────────────
SAVE to: docs/chapter6.md
REPORT: word count when done.


═══════════════════════════════════════════════════════════════
PROMPT 7 — Chapter 7: Implementation and Deployment
(Run after Chapter 6 is confirmed complete)
═══════════════════════════════════════════════════════════════

You are writing Chapter 7 of a diploma thesis.
THIS IS THE MOST IMPORTANT CHAPTER — worth 20 of 100 points.

BEFORE WRITING:
1. Read docs/codebase_map.md completely
2. Re-read: ALL handler files, the Dio client/interceptor file,
   ALL provider files, the WebSocket hub, docker-compose.yml,
   nginx config, CI/CD yml files, main router file

GRADING: 20 points. Grader expects: REST API documentation,
frontend-backend integration, WebSocket detail, feature deep-dives,
security implementation, Docker/CI/CD, testing.

THE GRADER PENALIZES: generic descriptions, missing specifics,
correct-sounding but inaccurate technical claims.

CONSTRAINTS:
- Minimum 3,500 words (this is the longest chapter)
- Reference actual file names, function names, endpoint paths
- Every claim must be verifiable in the codebase
- Include actual API endpoint documentation
- No LaTeX — plain Markdown

─────────────────────────────────────────────────────────────
# Chapter 7: Implementation and Deployment
─────────────────────────────────────────────────────────────

## 7.1 REST API Design and Endpoint Documentation

~500 words + complete endpoint table.

Explain the API design principles applied:
- Resource-based URL structure (/users, /matches, /messages, etc.)
- HTTP verb semantics (GET = read, POST = create, PATCH = update, DELETE = remove)
- Consistent JSON response envelope format
- JWT Bearer token authentication
- Pagination approach (limit/offset or cursor-based — check code)

Then document EVERY API endpoint found in the codebase:

| Module | Method | Path | Auth | Description |
|--------|--------|------|------|-------------|
| Auth | POST | /auth/register | No | Register new user |
| Auth | POST | /auth/login | No | Login, returns JWT |
| Auth | POST | /auth/refresh | Cookie | Refresh access token |
| [fill all from docs/codebase_map.md] | | | | |

After the table, show the standard response format:
{
  "status": "success",
  "data": { ... }
}
{
  "status": "error",
  "message": "reason"
}

## 7.2 Frontend-Backend Integration

~400 words.

Describe the complete integration architecture:

7.2.1 Dio HTTP Client Configuration
Show the DioClient setup: base URL, headers, timeout.
Describe both interceptors:
- Request interceptor: reads access_token from secure storage,
  injects Authorization: Bearer header
- Response error interceptor: catches 401, calls refresh endpoint,
  retries original request, handles refresh failure (logout)

7.2.2 Riverpod State Architecture
Describe how data flows from API to UI:
API response → repository parse → provider state update → widget rebuild

List key providers and their roles:
[Use exact provider names from codebase_map.md]

7.2.3 Error Handling Strategy
How network errors, auth errors, and domain errors are caught
and presented to the user.

## 7.3 WebSocket Implementation

~500 words — be very specific about the actual code.

7.3.1 Connection Establishment
- URL format: [exact URL from ApiConstants]
- How the Flutter client connects (WebSocketChannel)
- The CRITICAL auth handshake: first message must be:
  {"type": "auth", "token": "<JWT>"}
- What happens if auth message is missing or invalid
  (backend closes after 10s timeout)
- Why token is sent in first message, not URL param
  (URL params are logged; first message over TLS is secure)

7.3.2 Hub Architecture (Backend)
- How the Hub tracks connected clients (map of userID → connection)
- How incoming messages are routed to recipients
- How Redis pub/sub handles the case where recipient is on
  a different server instance
- The goroutine-per-connection model

7.3.3 Message Protocol
Document every message type with its JSON structure:
- Client → Server: auth, chat_msg, mahram_chat_msg, ping
- Server → Client: chat_msg, mahram_chat_msg, error, match_notification

7.3.4 Client-Side Message Handling (Flutter)
- How ChatNotifier manages connection state
- Optimistic local message insertion (why and how)
- isConnected state and UI feedback

## 7.4 Feature Implementation Deep-Dives

### 7.4.1 Discovery and Matching Engine
~350 words.
- The geolocation query: how distance is calculated (Haversine formula
  in SQL, or PostGIS — check the actual query in the repository)
- How age range filter is applied (min_age, max_age on birthdate)
- How niyyah filter works: only show users with matching niyyah
- How madhab filter works: show only same madhab or "any"
- How already-seen profiles are excluded (likes/passes table join)
- The like recording: what happens when user swipes right
- Mutual match detection: the atomic check-and-create operation
  that prevents duplicate matches

### 7.4.2 Community Feed
~200 words.
- Post creation with author info (how author_name is retrieved — JOIN)
- Feed pagination
- Like system (toggle like, like count)
- Comment system (nested or flat?)
- Real-time vs pull-to-refresh approach

### 7.4.3 Mahram Chat Feature
~250 words.
- What mahram means in Islamic context (guardian who supervises
  communication between unmarried people)
- How the mahram is added to a conversation (invitation flow)
- Technical implementation: how the mahram chat channel differs
  from regular chat in the Hub routing
- Access control: who can read the mahram channel
- How this is stored in the database

### 7.4.4 First Meeting Rating System
~200 words.
- The post-meeting review flow
- API endpoint: what fields are sent (check the handler)
- How the rating is stored and what it affects
- Data model for the rating (rated_user_id, rater_id, score, context)

### 7.4.5 Halal Settings System
~200 words.
- Niyyah filter: how it's stored (user profile) and applied (discovery query)
- Madhab filter: how it's stored and matched
- Modesty level: what it controls (photo visibility? matching pool?)
- Age range and distance settings: how they are persisted and used
- The settings PATCH endpoint and debounced save in Flutter

## 7.5 Security Implementation

~400 words.

7.5.1 JWT Architecture
- Algorithm: [from code — HS256 or RS256?]
- Access token: 15-minute expiry, contains userID and role claims
- Refresh token: 7-day expiry, stored as HttpOnly cookie
- Why HttpOnly cookie: prevents XSS access to refresh token
- Auth middleware: how it extracts and validates JWT from header
- Role-based access control: how supplier vs customer routes differ
  [adapt to this app's roles: regular user, mahram, admin]

7.5.2 Password Security
- bcrypt with cost factor [from code]
- Never storing plaintext or reversible hashes

7.5.3 Transport Security
- HTTPS everywhere via nginx TLS termination
- WebSocket over WSS (Secure WebSocket)
- CORS configuration: allowed origins, credentials: true

7.5.4 Rate Limiting
- nginx rate limiting rules [if found in config]
- Purpose: prevent brute force on auth endpoints

## 7.6 UI/UX Design and Screen Walkthrough

~300 words.

Design philosophy for a halal matrimony app:
- Modesty: design that respects Islamic sensibilities
- Trust: building confidence in the platform's Islamic compliance
- Simplicity: users range from tech-savvy to non-technical

Screen-by-screen walkthrough:
[Use actual screen files from docs/codebase_map.md]
For each screen: purpose, key UI elements, state management used,
navigation to/from

Key UX decisions to explain:
- Why onboarding collects niyyah and madhab before showing discovery
- How the matching card UI works (swipe mechanics)
- The chat UI: how mahram presence is indicated
- How the feed differs from a typical social feed

## 7.7 DevOps and Deployment

~450 words.

### 7.7.1 Docker and Containerization
- Describe the Dockerfile for the Go backend
  (base image, build stage, final stage, exposed port)
- Describe docker-compose.yml services:
  (backend, postgres, redis, nginx — with their configurations)
- Volume mounts for PostgreSQL persistence
- Environment variable management (.env files)
- Docker network for internal service communication

### 7.7.2 nginx Reverse Proxy
- Upstream definition for the Go backend
- Location blocks: /api/ → Go backend
- WebSocket proxy: the critical Upgrade header handling
  ("proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection 'upgrade';")
- SSL/TLS termination if configured
- Rate limiting zones

### 7.7.3 CI/CD Pipeline
[Describe based on what you found in .github/workflows/ or similar]
If no CI/CD found, describe what a proper pipeline would look like:
- Trigger: push to main branch
- Stage 1 — Build: go build, check compilation
- Stage 2 — Lint: golangci-lint, dart analyze
- Stage 3 — Test: go test ./..., flutter test
- Stage 4 — Docker: build and push image to registry
- Stage 5 — Deploy: SSH to server, docker-compose pull && up -d

### 7.7.4 Database Migrations
- Migration tool used [from code — golang-migrate, goose, or other]
- How migrations are run (at startup or manually)
- Migration file naming and versioning

## 7.8 Testing

~200 words.
- What tests exist in the backend (unit tests for services? handlers?)
- What tests exist in the Flutter frontend
- Key test scenarios: auth flow, match creation, message routing
- Manual testing approach for WebSocket flows
- Areas without test coverage (honest acknowledgment)

─────────────────────────────────────────────────────────────
SAVE to: docs/chapter7.md
REPORT: word count when done.


═══════════════════════════════════════════════════════════════
FINAL PROMPT — Assemble Complete Document
(Run only after ALL chapters are confirmed complete)
═══════════════════════════════════════════════════════════════

Your task: assemble all chapters into one complete diploma thesis.

READ in order:
  docs/codebase_map.md (for app name and key details)
  docs/chapter1.md
  docs/chapter2.md
  docs/chapter3.md
  docs/chapter4.md
  docs/chapter5.md
  docs/chapter6.md
  docs/chapter7.md
  docs/references.md

CREATE: docs/diploma_thesis_FINAL.md

Structure the final document as:

─── TITLE PAGE ───────────────────────────────────────────────
[App Name]
A Platform for Halal Matrimony and Islamic Matchmaking

By [student names — extract from Git history or leave as [Author Names]]
Department of Computer Engineering / Software Engineering
[University Name]
Astana, 2025

─── ABSTRACT ─────────────────────────────────────────────────
Write a 250-word abstract summarizing:
- What was built and why
- Key technical choices (Flutter, Go, PostgreSQL, Redis, WebSocket)
- Key Islamic domain features (niyyah, mahram chat, madhab filtering)
- Main outcomes and contributions

─── TABLE OF CONTENTS ────────────────────────────────────────
Generate a complete table of contents matching the document structure.
Include all chapters, sections, and subsections with their headings.
(Students will add page numbers manually in LaTeX.)

─── CHAPTERS 1–7 ─────────────────────────────────────────────
Paste chapters in order. Do NOT rewrite or summarize them.
Copy them verbatim from the chapter files.

─── REFERENCES ───────────────────────────────────────────────
Paste docs/references.md in full.
Verify all references are in IEEE format and numbered sequentially.

─── CONSISTENCY FIXES (apply while assembling) ───────────────
Before pasting each chapter, fix these if found:
- Inconsistent section numbers → fix to match TOC
- Chapter-internal cross-references ("as discussed in Chapter X")
  → verify they refer to the correct chapter
- Citation numbers → ensure they are globally sequential [1]-[N]
  not restarting at [1] in each chapter
- App name → ensure consistent spelling throughout

Do NOT:
- Rewrite or shorten any chapter content
- Add new content not in the chapter files
- Remove any sections

REPORT: "Final document assembled. Total word count: [N].
         Total citations: [N]. Saved to docs/diploma_thesis_FINAL.md"
```
