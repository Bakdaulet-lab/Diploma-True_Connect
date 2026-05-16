# МИНИСТЕРСТВО НАУКИ И ВЫСШЕГО ОБРАЗОВАНИЯ РЕСПУБЛИКИ КАЗАХСТАН

## Казахский национальный университет имени аль-Фараби

### Факультет информационных технологий

### Кафедра программной инженерии

---

&nbsp;

&nbsp;

# ДИПЛОМНАЯ РАБОТА

## **TrueConnect: разработка мобильного приложения для исламского бракосочетания с системой доверия на основе графовых алгоритмов**

&nbsp;

&nbsp;

**Специальность:** 6B06103 — Программная инженерия

**Выполнил(а):** Студент(ка) 4 курса

**Научный руководитель:** к.т.н., доцент кафедры программной инженерии

&nbsp;

&nbsp;

---

**Алматы — 2026**

---

&nbsp;

## АННОТАЦИЯ (ҚАЗАҚША)

Бұл дипломдық жұмыс Қазақстандағы мұсылман қауымдастығына арналған «TrueConnect» атты халал матримониялық мобильді қосымшаны жасауды және іске асыруды сипаттайды. Қосымша исламдық ниет (ниет), махрам (кіші жезде) қадағалауы, мазхаб үйлесімділігі және қауымдастыққа негізделген сенім ұпайы сияқты домендік талаптарды бірінші деңгейдегі архитектуралық шектеулер ретінде қарастырады. Жүйе Go 1.25 + Gin бэкендінде (таза архитектура), Flutter 3.16 мобильді клиентінде, PostgreSQL 16 + PostGIS реляциялық дерекқорында, Neo4j 5 графтік дерекқорында, Redis 7 кэштеу жүйесінде және MinIO объектілік қоймасында жазылған. Негізгі ерекшеліктері: Neo4j PageRank алгоритміне негізделген 0-100 балл сенім ұпайы жүйесі; AES-256-GCM шифрланған үш тараптық WebSocket чат арнасы; ниет мен мазхабты ескеретін геолокациялық іздеу алгоритмі; жеті сатылы исламдық некелесу жолы. 102+ бірлік тесті өтіп, толыққанды Docker Compose деплойы жасалды. Зерттеу бағдарламалық жасақтамада діни және этикалық талаптарды архитектуралық қорытынды ретінде қолданудың жаңа әдістемесін ұсынады.

**Кілт сөздер:** исламдық матримония, мобильді қосымша, сенім жүйесі, Neo4j, Flutter, Go, Clean Architecture, WebSocket, махрам, ниет.

---

## АННОТАЦИЯ (РУССКИЙ)

Данная дипломная работа описывает проектирование и реализацию мобильного приложения TrueConnect — халяльной брачной платформы для мусульманской общины Казахстана. Приложение рассматривает исламские доменные требования (ниет, надзор махрама, совместимость мазхабов, репутационная система) как архитектурные ограничения первого класса, а не как косметические функции. Серверная часть реализована на Go 1.25 + Gin с применением чистой архитектуры и предметно-ориентированного проектирования; мобильный клиент — на Flutter 3.16 (Riverpod, GoRouter, Dio). Хранилища данных: PostgreSQL 16 + PostGIS (реляционные данные и геолокация), Neo4j 5 Community + GDS (граф доверия, обнаружение Sybil-атак), Redis 7 (WebSocket pub/sub, кэш), MinIO (S3-совместимое хранилище фотографий и KYC-документов). Ключевые функции: система доверия 0–100 на основе взвешенного PageRank с байесовским сглаживанием; трёхсторонний зашифрованный чат с махрамом через WebSocket; геолокационный алгоритм подбора кандидатов с фильтрами ниет и мазхаба; семиэтапный исламский путь к никаху. Пройдено 102+ сервисных теста с детектором гонок. Все данные пользователей хранятся на казахстанских серверах в соответствии с требованиями законодательства РК о персональных данных.

**Ключевые слова:** исламское приложение для знакомств, мобильная разработка, Neo4j, граф доверия, Flutter, Go, чистая архитектура, WebSocket, Казахстан.

---

## ABSTRACT (ENGLISH)

This thesis presents the design and implementation of TrueConnect, a halal matrimony mobile application for the Muslim community in Kazakhstan. Unlike existing Islamic-branded dating platforms that apply superficial religious features to secular architectures, TrueConnect embeds Islamic domain requirements — niyyah (marital intention), mahram (guardian) supervision, madhab (jurisprudential school) compatibility, and community trust scoring — as first-class architectural constraints at every layer of the system. The backend is implemented in Go 1.25 with the Gin framework, using Clean Architecture with Domain-Driven Design across fifteen modules. The Flutter 3.16 mobile client uses Riverpod for state management and GoRouter for navigation. The data layer comprises PostgreSQL 16 with PostGIS (relational data and geospatial queries), Neo4j 5 Community with the Graph Data Science plugin (trust graph and Sybil detection), Redis 7 (WebSocket pub/sub fan-out and caching), and self-hosted MinIO (photo and KYC document storage). Key technical contributions include: a Bayesian-smoothed Neo4j PageRank trust score (0–100) with KYC verification weighting; an AES-256-GCM encrypted three-party WebSocket mahram supervision channel; a PostGIS-based candidate discovery algorithm with niyyah and madhab filters; and a seven-step structured Islamic courtship pathway from niyyah declaration to nikah confirmation. The system passes 102+ service-layer tests under Go's race detector and is deployed as a complete Docker Compose stack on Kazakhstani infrastructure, satisfying the Republic of Kazakhstan's data sovereignty requirements.

**Keywords:** Islamic matrimony application, mobile development, trust scoring, Neo4j graph algorithms, Flutter, Go, Clean Architecture, WebSocket, halal, Kazakhstan.

---

&nbsp;

## TABLE OF CONTENTS

- [Abstract (Kazakh)](#аннотация-қазақша)
- [Abstract (Russian)](#аннотация-русский)
- [Abstract (English)](#abstract-english)
- [Chapter 1: Introduction](#chapter-1-introduction)
  - [1.1 Background and Motivation](#11-background-and-motivation)
  - [1.2 Problem Statement](#12-problem-statement)
  - [1.3 Significance and Impact](#13-significance-and-impact)
  - [1.4 Research Objectives](#14-research-objectives)
  - [1.5 Scope and Limitations](#15-scope-and-limitations)
  - [1.6 Structure of the Document](#16-structure-of-the-document)
- [Chapter 2: Literature Review](#chapter-2-literature-review)
  - [2.1 Online Dating Platforms: Evolution and Behavioural Research](#21-online-dating-platforms-evolution-and-behavioural-research)
  - [2.2 Islamic Marriage Practices and the Role of Technology](#22-islamic-marriage-practices-and-the-role-of-technology)
  - [2.3 Recommendation and Matching Algorithms](#23-recommendation-and-matching-algorithms)
  - [2.4 Real-Time Communication Architecture](#24-real-time-communication-architecture)
  - [2.5 Software Architecture for Complex Domains](#25-software-architecture-for-complex-domains)
  - [2.6 Cross-Platform Mobile Development](#26-cross-platform-mobile-development)
  - [2.7 Literature Review Summary](#27-literature-review-summary)
- [Chapter 3: Analysis of Existing Systems](#chapter-3-analysis-of-existing-systems)
  - [3.1 Overview of the Competitive Landscape](#31-overview-of-the-competitive-landscape)
  - [3.2 Analysis of Individual Systems](#32-analysis-of-individual-systems)
  - [3.3 Functional Comparison Table](#33-functional-comparison-table)
  - [3.4 SWOT Analysis](#34-swot-analysis)
  - [3.5 Gap Analysis](#35-gap-analysis)
  - [3.6 Data Collection and User Research Methodology](#36-data-collection-and-user-research-methodology)
  - [3.7 Challenges Unique to Islamic Matrimony Applications](#37-challenges-unique-to-islamic-matrimony-applications)
- [Chapter 4: Methodology](#chapter-4-methodology)
  - [4.1 Development Methodology: Agile with Scrum](#41-development-methodology-agile-with-scrum)
  - [4.2 Domain-Driven Design Application](#42-domain-driven-design-application)
  - [4.3 Requirements Engineering](#43-requirements-engineering)
  - [4.4 Testing Strategy](#44-testing-strategy)
  - [4.5 Justification of Methodology](#45-justification-of-methodology)
- [Chapter 5: System Architecture and Design](#chapter-5-system-architecture-and-design)
  - [5.1 System Architecture Overview](#51-system-architecture-overview)
  - [5.2 Backend Clean Architecture Layers](#52-backend-clean-architecture-layers)
  - [5.3 Backend Module Architecture](#53-backend-module-architecture)
  - [5.4 Database Architecture](#54-database-architecture)
  - [5.5 WebSocket Architecture](#55-websocket-architecture)
  - [5.6 UML Diagram Specifications](#56-uml-diagram-specifications)
- [Chapter 6: Technology Selection and Justification](#chapter-6-technology-selection-and-justification)
  - [6.1 Backend Language and Runtime](#61-backend-language-and-runtime)
  - [6.2 Mobile Framework](#62-mobile-framework)
  - [6.3 Primary Relational Database](#63-primary-relational-database)
  - [6.4 Social Graph Database](#64-social-graph-database)
  - [6.5 Real-Time Communication Protocol](#65-real-time-communication-protocol)
  - [6.6 Caching and Session Store](#66-caching-and-session-store)
  - [6.7 Object Storage](#67-object-storage)
  - [6.8 Summary](#68-summary)
- [Chapter 7: Implementation and Deployment](#chapter-7-implementation-and-deployment)
  - [7.1 Overview](#71-overview)
  - [7.2 Backend Implementation](#72-backend-implementation)
  - [7.3 Frontend Implementation](#73-frontend-implementation)
  - [7.4 Security Implementation](#74-security-implementation)
  - [7.5 Deployment Pipeline](#75-deployment-pipeline)
  - [7.6 Testing Summary](#76-testing-summary)
  - [7.7 Demo Walkthrough](#77-demo-walkthrough)
- [References](#references)

---

# Chapter 1: Introduction

## 1.1 Background and Motivation

The institution of marriage holds a position of profound importance in Islamic theology and jurisprudence. The Prophet Muhammad (peace be upon him) described marriage as completing half of one's faith, and the Quran (30:21) explicitly frames it as a source of tranquillity, love, and compassion — values that stand in sharp contrast to the transactional and gamified nature of contemporary digital matchmaking. With an estimated 1.8 billion Muslims worldwide, representing approximately 24 percent of the global population, the challenge of finding a spouse who aligns with Islamic values in an increasingly digital society constitutes a significant social and technical problem that has yet to receive an adequate engineering solution.

The proliferation of mainstream dating applications — most notably Tinder, Bumble, and Hinge — has fundamentally restructured how people form romantic connections in the modern era. These platforms share a common design philosophy: a photograph-first interface optimised for rapid, low-friction evaluation of potential partners, a swipe-based gamification mechanic that incentivises volume over intentionality, and an implicit assumption of casual or exploratory intent. From an Islamic perspective, each of these design choices directly contravenes established religious principles. The unrestricted mixing of unrelated men and women (known as khalwa in Arabic), the emphasis on physical appearance over character and piety, and the absence of any declaration of serious marital intent (niyyah) make these platforms fundamentally incompatible with Islamic courtship norms.

Several applications have positioned themselves as Islamic alternatives to mainstream dating platforms. Muzz (formerly Muzmatch), Salams (formerly Minder), and Hawaya by Match Group represent the most prominent offerings in this space. However, a technical and functional analysis of these platforms reveals that they are predominantly adaptations of Western dating application patterns with superficial Islamic branding rather than platforms engineered around Islamic domain requirements. None of these applications offer a functional mahram supervision system — the Islamic concept of a guardian (محرم) whose presence is required during interactions between unmarried individuals. None support filtering of prospective matches by madhab (مذهب, the Islamic school of jurisprudential thought, of which the four primary schools are Hanafi, Shafi'i, Maliki, and Hanbali), which is a meaningful compatibility dimension for observant Muslims. Furthermore, none implement niyyah (نية) as a formal first-class concept that shapes the algorithmic matching behaviour of the platform.

The specific context of Kazakhstan and Central Asia intensifies both the need and the opportunity. Kazakhstan has a Muslim-majority population of approximately 70 percent, a rapidly growing smartphone penetration rate, and a cultural tradition of community-mediated marriage processes. Yet no locally developed, locally hosted Islamic matrimony platform exists for this region. This absence means that Kazakhstani Muslim users who seek digital matchmaking assistance must either use platforms that violate their religious values, or platforms that are designed for Middle Eastern or South Asian cultural contexts that differ significantly from the Central Asian Muslim experience.

The technical argument for this project is equally compelling. Existing Islamic dating applications treat religious requirements as user interface elements — profile badges, modesty toggles — rather than as architectural constraints. TrueConnect was designed from first principles with Islamic domain concepts embedded at every layer of the system: the domain model (`internal/domain/`), the database schema (`social.profiles.niyyah`, `social.profiles.madhab`, `social.mahrams`), the matching algorithm (niyyah compatibility filtering, madhab affinity scoring), the communication architecture (the mahram chat room system), and the user interface (the `NiyyahSelectionScreen` presented during onboarding before any discovery begins). This architectural philosophy — treating Islamic values as first-class technical requirements rather than cosmetic additions — is the defining contribution of this project.

---

## 1.2 Problem Statement

A systematic examination of the current landscape of digital matrimony solutions, combined with an analysis of the technical and social requirements of Muslim users in Kazakhstan, reveals five discrete and well-defined problems that the existing generation of applications fails to address.

**Problem 1: The absence of mahram (guardian) supervision in digital communication channels.** Islamic jurisprudence across all four major schools of thought holds that private communication between an unmarried man and an unrelated woman, without the presence of a mahram, is impermissible. No existing mainstream or Islamic-branded dating application provides a genuine technical implementation of this requirement. Existing platforms, at best, provide a voluntary "wali awareness" notification — informing a designated guardian that a woman is using the platform — which bears no resemblance to actual supervised communication. The result is that observant Muslim women face an irreconcilable conflict between using digital matrimony tools and adhering to their religious obligations.

**Problem 2: The inability to filter prospective matches by madhab (Islamic school of jurisprudence).** Madhab compatibility is a meaningful and practically important dimension of marital compatibility for observant Muslims. A Hanafi Muslim and a Hanbali Muslim may have significant differences in their daily religious practice, family rituals, and dietary norms that affect long-term marital harmony. No existing platform surfaces or filters on madhab as a matching criterion. This gap forces users to discover this incompatibility only after extensive communication, wasting significant time and emotional investment.

**Problem 3: No formal declaration of niyyah (marital intention) that shapes algorithmic behaviour.** Mainstream platforms implicitly assume casual or exploratory intent. Even Islamic-branded applications that include a "niyyah" profile field treat it as a display label rather than a parameter that governs who appears in a user's discovery feed. This means that a user seeking serious marriage within the year (nikah_year) is shown candidates who are interested only in friendship, and vice versa — a mismatch that degrades the quality of matches and potentially exposes users to interactions that violate their stated intentions.

**Problem 4: Mainstream applications incentivise superficial, high-volume swiping behaviour that is antithetical to Islamic courtship values.** The gamification mechanics of swipe-based discovery — unlimited likes, visual-first card design, badge counts — are engineered to maximise engagement time and swipe volume. This design philosophy is fundamentally at odds with the Islamic emphasis on deliberate, intention-driven spouse selection guided by character, piety, and compatibility rather than physical appearance. A platform designed for Muslim users must create friction that encourages reflection, not mechanics that reward volume.

**Problem 5: No trust and reputation infrastructure that reflects real-world character and Islamic integrity.** Islamic marriage tradition places significant weight on the testimony of others regarding a person's character (the role of the wali and community witnesses). Digital platforms offer no equivalent mechanism. Fake profiles, misrepresentation, and Sybil attacks undermine user safety and trust in ways that are especially harmful in a matrimony context, where users are making life-altering decisions. The absence of a verifiable, community-grounded reputation system leaves Muslim users without a digital equivalent of the community-based character verification that traditional Islamic matchmaking relies upon.

In summary, the core problem addressed by this thesis is the complete absence of a mobile matrimony platform that treats Islamic domain requirements — mahram supervision, madhab compatibility, niyyah filtering, intentional courtship mechanics, and community-grounded trust scoring — as first-class architectural requirements rather than cosmetic additions. TrueConnect was designed and implemented to solve this problem specifically for the Kazakhstani Muslim community.

---

## 1.3 Significance and Impact

The significance of TrueConnect operates across four distinct dimensions: social, technical, economic, and research.

**Social Impact.** For Muslim users in Kazakhstan, the existence of a platform that natively enforces Islamic courtship norms has concrete, meaningful implications for their daily lives. A Kazakh Muslim woman using TrueConnect can communicate with potential marriage candidates knowing that a mahram is technically present in every conversation — not as a courtesy notification, but as an active participant in a three-way encrypted chat channel (`social.mahram_chat_rooms`). She can filter discovery results by niyyah and madhab without manual post-hoc screening. She can verify candidates through a community trust score that is grounded in real-world interactions and KYC identity verification, rather than self-reported claims.

For Muslim men, the platform provides symmetric benefits: a candidate pool pre-filtered for serious marital intent, a trust score system that rewards genuine character rather than profile attractiveness, and a structured courtship pathway — niyyah declaration → profile view → match → supervised chat → family introduction → imam connection → nikah confirmation — that mirrors the Islamic courtship process rather than subverting it.

**Technical Significance.** From a software engineering perspective, TrueConnect makes a contribution to the emerging field of values-aligned application design. The architecture demonstrates a concrete methodology for embedding domain-specific ethical and religious requirements at every layer of a production system: domain model (`internal/domain/profile.go` with niyyah and madhab as typed enums), database schema (the `social.mahrams` table and `identity_vault` dual-schema architecture), matching algorithm (niyyah compatibility filtering in `internal/service/matching_service.go`), and user interface (the `NiyyahSelectionScreen` as a mandatory onboarding gate before discovery is accessible). This methodology is transferable to other domains where ethical or religious constraints must be technically enforced rather than optionally declared.

**Economic Opportunity.** The Islamic digital economy is a rapidly growing sector. The halal industry globally is estimated at over USD 2 trillion, yet digital services — particularly social platforms — represent a disproportionately small share of this market. Within Kazakhstan specifically, no commercially successful locally-hosted Islamic matrimony platform exists. TrueConnect occupies an entirely uncontested market position.

**Research Contribution.** This project contributes to academic knowledge in three specific areas: (1) the application of Clean Architecture and Domain-Driven Design to an Islamic social domain model; (2) the design and implementation of a community-grounded trust score system using Neo4j graph algorithms (weighted PageRank with Bayesian smoothing) for a social matchmaking context; and (3) the engineering of a three-party supervised real-time communication channel using WebSockets and Redis pub/sub that preserves AES-256-GCM encryption across all participants.

---

## 1.4 Research Objectives

1. **To design and implement a niyyah-based matching algorithm** using PostgreSQL and Redis in order to ensure that users with incompatible marital intentions are never shown to each other in the discovery feed.

2. **To build a three-way encrypted mahram chat channel** using WebSockets, AES-256-GCM symmetric encryption, and Redis pub/sub in order to provide a technically enforced Islamic guardian supervision mechanism.

3. **To develop a community-grounded trust and reputation scoring engine** using Neo4j graph database algorithms (weighted Bayesian scoring with KYC verification multipliers).

4. **To implement a complete Islamic courtship pathway** as a structured sequence of verifiable milestones — niyyah declaration, match formation, supervised chat, family introduction, imam connection, and nikah confirmation.

5. **To engineer a privacy-preserving identity verification system** using a dual-schema PostgreSQL architecture (`social` + `identity_vault`) with AES-256-GCM field-level encryption and Argon2id hashing.

6. **To design and implement a modesty-compliant discovery interface** using Flutter and the `no_photo_mode` profile feature in order to give female users granular control over their photo visibility.

7. **To build a scalable, self-hosted mobile backend** using Go 1.25 + Gin, a modular monolith Clean Architecture, Docker Compose orchestration (PostgreSQL 16 + PostGIS, Neo4j 5, Redis 7, MinIO), and an nginx reverse proxy.

8. **To develop and validate a comprehensive test suite** covering all service-layer business logic using Go's testing package with the `-race` detector, achieving 102+ passing tests.

---

## 1.5 Scope and Limitations

**In Scope.** User registration and authentication (JWT HS256, Argon2id, AES-256-GCM PII encryption); profile creation and management with Islamic domain fields (niyyah, madhab, languages, no-photo mode, marital status); geolocation-based discovery (PostGIS `ST_DWithin`); niyyah and madhab filtering; swipe mechanics; mutual match detection; real-time WebSocket chat with AES-256-GCM message encryption and Redis pub/sub; mahram chat rooms; community feed; interaction ratings with trust score recalculation; KYC identity submission; notifications with FCM; imam directory (10 imams, 5 KZ cities); whisper reporting; admin panel; and a complete Flutter application across 21 screens.

**Out of Scope.** Phone number OTP verification; video/audio calling beyond WebRTC signaling; Next.js web frontend; payment processing; full APNs iOS provisioning.

**Current Limitations.** The trust score event channel (capacity 100) can drop events under extreme load. The Neo4j advisory lock for `RecordLike` has not been implemented, leaving a theoretical race condition under simultaneous identical like requests. The imam catalog requires binary redeployment to update. iOS testing was limited by the absence of Apple Developer enrollment.

---

## 1.6 Structure of the Document

**Chapter 1** establishes motivation, problem statement, significance, research objectives, and scope.

**Chapter 2** surveys academic literature across six areas: online dating behaviour, Islamic marriage practices, matching algorithms, real-time communication, software architecture, and cross-platform mobile development.

**Chapter 3** provides competitive analysis of five existing platforms (Tinder, Bumble, Muzz, Salams, Hawaya), a 20-row functional comparison table, SWOT analysis, and gap analysis mapping each deficiency to its codebase implementation.

**Chapter 4** explains the Agile development methodology (14 sprints), DDD application (bounded contexts, ubiquitous language, aggregates), requirements engineering (22 functional + 12 non-functional requirements), and testing strategy.

**Chapter 5** presents the complete technical architecture: three-tier overview, five Clean Architecture layers, all backend modules, database design, WebSocket Hub architecture, and eight UML diagram specifications.

**Chapter 6** presents comparative technology analyses: Go vs. Node.js vs. Python; Flutter vs. React Native vs. native; PostgreSQL vs. MongoDB vs. Firebase; WebSocket vs. alternatives; Neo4j vs. PostgreSQL recursive CTEs; Redis vs. Memcached.

**Chapter 7** documents the complete implementation: backend API, security, frontend screens, deployment pipeline, CI/CD, and a reproducible demo walkthrough.

---

# Chapter 2: Literature Review

## 2.1 Online Dating Platforms: Evolution and Behavioural Research

The academic study of online dating has matured considerably since the first wave of empirical research in the mid-2000s, evolving from descriptive accounts of early platform adoption to sophisticated analyses of algorithmic influence on mate selection, user self-presentation strategies, and the psychological consequences of gamified matchmaking. Understanding this literature is essential context for TrueConnect, because the platform is in many respects a deliberate architectural counter-argument to the design patterns that empirical research has identified as dominant in the industry.

The foundational comprehensive review of online dating science was conducted by Finkel, Eastwick, Karney, Reis, and Sprecher [1], whose analysis in Psychological Science in the Public Interest remains the most cited critical treatment of the field. The authors identify three categories of claims made by online dating platforms: that computer-mediated communication is superior to face-to-face communication for initial partner evaluation; that access to a large pool of potential partners improves matching outcomes; and that mathematical compatibility algorithms improve matching accuracy. Their systematic review finds the empirical support for each of these claims to be weaker than platform marketing suggests. Crucially for the TrueConnect design, Finkel et al. note that algorithmic matching systems tend to reduce potential partners to sets of profile attributes and thereby fail to capture the interactive chemistry that predicts long-term relationship satisfaction. This finding reinforces the TrueConnect design decision to use constraint-based niyyah and madhab filtering not as a similarity optimisation but as a requirement filter — ensuring that users are never shown candidates with fundamentally incompatible intentions — while leaving the qualitative dimensions of compatibility to emerge through supervised conversation.

The analysis of actual user behaviour on contemporary swipe-based platforms by Tyson, Perta, Haddadi, and Seto [2] provides quantitative evidence of the pathological dynamics that gamified discovery mechanics create. Analysing Tinder data across multiple metropolitan areas, they find extreme asymmetry in like rates between male and female users and demonstrate that the swipe mechanic creates behaviour more closely resembling slot-machine engagement than deliberate partner evaluation. Male users swipe right on approximately 46 percent of profiles, while female users approve approximately 14 percent — a disparity that the authors argue is a direct consequence of the interface design rather than underlying preference differences. TrueConnect's discovery screen intentionally preserves the familiar swipe metaphor while removing the like-volume incentive: there are no like counters, no match-percentage badges, and no "super like" mechanics.

The economics of online partner matching were rigorously analysed by Hitsch, Hortaçsu, and Ariely [3] in a study using data from a major US online dating platform. Their analysis demonstrates that revealed preferences in partner selection are strongly driven by income, physical attractiveness, and height — attributes that have limited relevance in Islamic spouse selection where piety, family background, madhab compatibility, and declared intention are given priority. This divergence between the attribute dimensions rewarded by mainstream matching optimisation and the attribute dimensions prioritised in Islamic courtship is a central argument for building a domain-specific platform rather than customising an existing one.

Self-presentation dynamics in online dating profiles have been extensively studied. Ellison, Heino, and Gibbs [4] find that users engage in a process of "selective self-presentation" — revealing information strategically to manage impressions while attempting to maintain plausibility for eventual face-to-face encounters. This creates a tension between optimistic self-presentation and the Islamic prohibition on deception (ghish). Toma, Hancock, and Ellison [5] extend this analysis with objective measurements, finding that male users systematically overstate height and income while female users understate weight. TrueConnect's KYC identity verification system (`social.kyc_submissions`, `identity_vault.iin_vault`) directly addresses this deception dynamic by providing a verifiable identity layer: a KYC-verified badge (communicated via `verification_level` in `social.users`) signals to other users that the profile has been cross-checked against a government identity document.

---

## 2.2 Islamic Marriage Practices and the Role of Technology

Islamic marriage is governed by a rich and well-documented jurisprudential tradition that prescribes specific roles, processes, and ethical constraints absent from secular matchmaking. Engagement with this literature is necessary to justify the specific domain design decisions made in TrueConnect.

The global Muslim population context is provided by the Pew Research Center's comprehensive demographic analysis [6], which projects that the Muslim population will reach 2.76 billion by 2050, representing 29.7 percent of the global total. Kazakhstan is specifically identified as a Muslim-majority country (approximately 70 percent of the population identifying as Muslim), reinforcing the market rationale for a locally developed platform.

The broader relationship between Islam and digital technology has been studied by Bunt [7], whose work on "iMuslims" documents the extensive and sophisticated use of digital networks by Muslim communities for religious learning, community organisation, and social connection. Bunt argues that digital Islam is not a dilution of traditional practice but an extension of the Islamic tradition of adapting available technology to religious purposes. This framing legitimises the TrueConnect project as a continuation of a documented pattern of Islamic digital practice rather than a novelty or cultural compromise.

The specific challenges of identity and religious performance in online spaces for Muslim women have been studied by Kavakci and Kraeplin [8], who document the strategies used by observant Muslim women to maintain religious identity in digital social environments not designed for their needs. Their work directly informs the TrueConnect `no_photo_mode` feature: rather than forcing observant women to choose between visibility and modesty, TrueConnect treats modesty control as an architectural requirement. When `no_photo_mode = true` in `social.profiles`, the discovery query returns `AvatarURL = ""` and sets `NoPhotoMode = true` in the `CandidateRow` struct, which the Flutter `DiscoveryScreen` renders as a blurred placeholder.

The ethics of digital communication in Islamic contexts have been addressed by Al-Saggaf [9], who argues that Islamic online communities require specific ethical infrastructure — mechanisms for accountability, truthfulness, and guardianship — that secular platforms do not provide. This argument directly maps to the three technical innovations in TrueConnect that have no parallel in secular dating applications: the mahram chat room system (guardian accountability), the KYC identity verification (truthfulness infrastructure), and the trust score engine (community-based reputation).

---

## 2.3 Recommendation and Matching Algorithms

The algorithmic core of any matchmaking platform is its candidate recommendation and ranking system. The most comprehensive survey of recommender system approaches is provided by Bobadilla, Ortega, Hernando, and Gutiérrez [10], who classify recommender systems into three main paradigms: collaborative filtering, content-based filtering, and hybrid systems. In the matrimony domain, collaborative filtering would mean surfacing candidates whom similar users have liked — a mechanism that is both ethically problematic and practically counterproductive for the Islamic use case where individual religious compliance, family background, and declared intention are not capturable in aggregate similarity metrics.

TrueConnect's matching algorithm implements a form of constraint-based content filtering: the `FindCandidates` query applies hard constraint filters (niyyah compatibility via `AllowedNiyyahs`, block exclusion, seen-set exclusion, geographic distance via PostGIS `ST_DWithin`) and then applies a soft preference boost (madhab affinity +10 to `TrustScore`) to rank candidates within the constraint-satisfying set.

The theoretical foundation for stable two-sided matching was established by Gale and Shapley [11] in their seminal paper on college admissions, which introduced the concept of a stable matching — a pairing where no two unmatched individuals both prefer each other to their current assignment. TrueConnect's like/pass mechanism implements a version of this model: two users must mutually express interest before a match is created, ensuring that no match exists where one party is paired with someone they have explicitly rejected.

Hybrid recommender systems have been surveyed by Burke [12], who identifies seven distinct hybridisation strategies for combining content-based and collaborative approaches. The weighted hybrid most closely describes TrueConnect's graph-candidate system (`GET /v1/matching/graph-candidates`), which uses Neo4j PageRank-weighted trust scores to surface candidates who are highly regarded by the community while satisfying content-based filters.

Location-based social network (LBSN) recommendation is directly relevant to TrueConnect's geolocation-based discovery. Ye, Yin, Lee, and Lee [13] demonstrate that incorporating geographic influence into collaborative filtering significantly improves recommendation accuracy for location-sensitive social applications. TrueConnect's PostGIS-based distance query (filtering by `max_distance_km` from `social.user_settings`) implements the geographic constraint dimension of this research, ensuring that candidates are physically reachable for the in-person meetings that the Islamic courtship process requires.

---

## 2.4 Real-Time Communication Architecture

The WebSocket protocol, defined in IETF RFC 6455 by Fette and Melnikov [14], provides a full-duplex communication channel over a single TCP connection, fundamentally distinguishing it from the HTTP request-response paradigm. For TrueConnect's chat system, WebSocket is the only technically appropriate choice: the mahram chat requirement mandates that a message sent by any of three participants (woman, man, guardian) is delivered to all three simultaneously, which requires genuine server-push capability that HTTP polling and Server-Sent Events cannot provide without significant complexity and latency overhead.

The REST architectural style, described by Fielding and Taylor [15], is used for all non-real-time operations in TrueConnect (profile management, match creation, feed, notifications). The clean separation between REST API operations and WebSocket operations reflects the principle that architectural styles should be matched to their appropriate communication patterns rather than applied uniformly across all system interactions.

Redis as a messaging backbone for real-time systems is described in practical detail by Carlson [16]. TrueConnect uses Redis pub/sub with a per-match channel naming scheme (`chat:<matchID>`) to enable horizontal scaling: when a message is sent via WebSocket from User A to User B, the Hub first attempts direct delivery to User B's local connection. If User B is connected to a different server instance, the message is PUBLISHED to the Redis channel, which is SUBSCRIBED by all Hub instances, allowing the correct instance to deliver the message.

---

## 2.5 Software Architecture for Complex Domains

Clean Architecture, as articulated by Robert C. Martin [17], prescribes a concentric layer model in which the innermost layer contains pure business entities with no framework dependencies, and each outer layer depends only on inner layers — never the reverse. In TrueConnect, this manifests as the dependency rule: `internal/domain/` has zero external imports; `internal/service/` depends only on `internal/domain/` and `internal/repository/` interfaces; `internal/handler/` depends on services; and `internal/adapter/` implements repository interfaces but is never imported by services.

Domain-Driven Design (DDD), introduced by Evans [18], provides the vocabulary and structuring principles that allow a complex domain to be faithfully reflected in code. The central DDD concept of ubiquitous language is visibly implemented throughout TrueConnect. The Islamic domain terms `niyyah`, `mahram`, and `madhab` appear identically in Go struct field names, PostgreSQL column names, API endpoint paths, Flutter model fields, and screen route names.

The modular monolith pattern occupies an important position in the microservices literature as the appropriate intermediate architecture for systems that need clear module boundaries but do not yet have the operational complexity to justify microservice deployment. Newman [19] argues that a modular monolith is frequently the correct architectural choice for a system in its early growth phase: it preserves the option to extract modules into independent services later while avoiding the distributed systems complexity that microservices introduce from day one.

Fowler and Lewis [20] provide the canonical definition of microservices and the criteria by which an organisation should consider transitioning from a monolith. Their "micro" criteria — independently deployable, organised around business capabilities, built by small teams — would apply to TrueConnect at a future scale but are not met by the current project.

---

## 2.6 Cross-Platform Mobile Development

Biørn-Hansen, Majchrzak, and Grønli [21] provide a systematic comparison of progressive web applications and native mobile applications across multiple criteria, concluding that the choice of mobile development approach should be driven by the performance and user experience requirements of the specific application. Their framework, applied to TrueConnect, yields a clear recommendation toward compiled native or near-native approaches: the swipe-based discovery interface requires smooth 60fps card animations; the mahram chat screen requires real-time message delivery with sub-200ms visual update; and the trust score badge requires smooth numerical animation.

Flutter, as documented by Google [22], achieves near-native performance through a key architectural distinction: unlike React Native, which renders using the platform's native UI components via a JavaScript bridge, Flutter uses its own rendering engine (Skia/Impeller) that compiles directly to native ARM code. This means that Flutter animations run in the Dart VM with direct access to the GPU, without the latency overhead of a JavaScript bridge.

Nawrocki, Wrona, Marczak, and Szmeja [23] compare the architectural characteristics of cross-platform frameworks and find that type-safe compile-time dependency injection — which Riverpod provides through its `Provider` declarations and `ConsumerWidget` binding — is particularly valuable for applications with complex interdependent state, such as an app where authentication state must gate access to discovery, chat, and settings state simultaneously.

---

## 2.7 Literature Review Summary

The literature surveyed in this chapter converges on several conclusions that directly validate the design decisions made in TrueConnect. Empirical research on online dating behaviour [1, 2, 3] establishes that existing platforms' gamification mechanics and photograph-first design produce outcomes misaligned with the needs of users seeking serious, values-aligned partners. Islamic marriage scholarship [6, 7, 8, 9] confirms that Muslim users require specific technical infrastructure — guardian supervision, identity verification, deception prevention — that no existing platform provides. The recommender systems literature [10, 11, 12, 13] justifies TrueConnect's constraint-based filtering approach over collaborative filtering for the Islamic domain. The real-time communication literature [14, 15, 16] supports the WebSocket + Redis pub/sub architecture as the correct technical choice for the mahram three-party supervised chat requirement. The software architecture literature [17, 18, 19, 20] validates the Clean Architecture + DDD + modular monolith pattern as appropriate for a system of this complexity. The mobile development literature [21, 22, 23] confirms Flutter as the correct platform choice given TrueConnect's animation and state management requirements.

The specific gap that this project addresses — the absence of a production-quality, self-hosted mobile platform that treats Islamic domain requirements as first-class architectural constraints — is filled by TrueConnect through the integration of technical choices drawn from each of the literature areas surveyed, unified by a domain model that places the Islamic courtship process at the centre of every design decision.

---

# Chapter 3: Analysis of Existing Systems

## 3.1 Overview of the Competitive Landscape

The landscape of digital matchmaking applications can be organised into three distinct categories. The first consists of mainstream dating applications — principally Tinder and Bumble — designed for secular Western markets. The second consists of Islamic-branded matrimony applications — primarily Muzz, Salams, and Hawaya — that have attempted to capture the Muslim user segment by applying Islamic branding to architectures inherited from the mainstream category. The third consists of regional and traditional matchmaking infrastructure specific to Central Asia and Kazakhstan, which operates largely offline through family networks and community intermediaries.

---

## 3.2 Analysis of Individual Systems

### 3.2.1 Tinder

Tinder, launched by Match Group in 2012, is the world's most widely used dating application with over 75 million monthly active users. Its core mechanic — swiping right to express interest, left to decline — reduced the cognitive overhead of partner evaluation to a binary gesture applied to a photograph. From an Islamic compliance perspective, Tinder exhibits fundamental structural incompatibilities: the platform is photograph-first by architecture; there is no mechanism to declare marital intention, no concept of guardian supervision, no filtering by religious school of thought, and no reputation or character verification. Tinder's engagement metrics are optimised for sustained swiping activity rather than successful long-term matching — a documented conflict of interest that Tyson et al. [2] quantify in their behavioural analysis.

### 3.2.2 Bumble

Bumble, founded in 2014, differentiates itself through its "women-first" messaging rule: after a match is created, only the woman can initiate the first message within 24 hours. While Bumble's women-first mechanic superficially resembles the Islamic principle of female agency in marriage consent, the structural similarities end there. Like Tinder, Bumble is photograph-first, intention-agnostic, and has no guardian supervision mechanism. The 24-hour message expiry creates time pressure antithetical to the deliberate, family-mediated pace of Islamic courtship. Bumble does not address any of the five problems identified in Chapter 1.

### 3.2.3 Muzz (formerly Muzmatch)

Muzz, with over 10 million registered users, includes several features marketed as Islamic-compliant: a "chaperone" mode that copies a designated third party on all messages, a modesty filter that blurs profile photos, and a niyyah field on profiles. However, a technical analysis reveals that these Islamic features are implemented as user-interface overlays rather than architectural constraints. The chaperone mode is implemented as a carbon-copy email notification — the third party receives email copies of messages but is not an active participant in the conversation channel and cannot send messages. This is categorically different from TrueConnect's mahram chat room system, where the guardian is an authenticated participant in a three-way encrypted WebSocket channel with full send/receive capability. Muzz's niyyah field does not filter the discovery algorithm.

### 3.2.4 Salams (formerly Minder)

Salams, with approximately 4 million users, closely resembles Tinder in its UX: card-based swipe discovery, photo-first profiles, and bilateral like matching. Salams includes some Islamic profile fields — ethnicity, religiosity level, and prayer frequency — but these are display attributes that do not influence the matching algorithm's behaviour. There is no mahram chat feature, no niyyah filter, no madhab preference filtering, and no trust or reputation system. Salams' "Premium" subscription tier gates basic features such as seeing who liked you — a business model contrary to the Islamic principle that access to spouse-finding should not be contingent on financial capacity.

### 3.2.5 Hawaya (by Match Group)

Hawaya, launched in 2019 by Match Group, targets Arab Muslim women with a "Guardian Mode" — which is an email notification mechanism rather than a technical supervision system. The fundamental limitation of Hawaya is structural: it is a Match Group product built on Match Group infrastructure, meaning all user data is stored under US jurisdiction. For a matrimony application handling sensitive personal information for Muslim users in Kazakhstan, US data jurisdiction is a critical concern given Kazakhstan's data localisation legislation. Hawaya has no Kazakh or Russian language support, no regional imam directory, and no features reflecting Central Asian Islamic traditions.

### 3.2.6 Regional and Traditional Matchmaking Platforms

The traditional matchmaking infrastructure in Kazakhstan operates primarily through family and community networks, professional matchmakers ("свахи" / "қасиетті жеңге"), and informal groups on WhatsApp and Telegram. Several Kazakhstani websites (nikah.kz, secondhalf.kz) attempt to provide digital platforms for this need, but these are web-only services with no native mobile applications, no real-time communication features, no algorithmic matching, and no trust or reputation infrastructure. The absence of any full-featured, locally hosted, halal-compliant mobile matrimony application in Kazakhstan represents the precise market gap that TrueConnect was designed to fill.

---

## 3.3 Functional Comparison Table

| Feature | Tinder | Bumble | Muzz | Salams | Hawaya | **TrueConnect** |
|---|---|---|---|---|---|---|
| Mahram / guardian chat supervision | No | No | Partial (email CC only) | No | Partial (email CC only) | **Yes** (3-way WS channel) |
| Niyyah declaration shapes algorithm | No | No | No (display only) | No | No | **Yes** (`AllowedNiyyahs` filter) |
| Madhab school filtering | No | No | No | No | No | **Yes** (+10 affinity boost) |
| Photo modesty / NoPhotoMode | No | No | Partial (blur toggle) | No | No | **Yes** (architectural enforcement) |
| Community trust / reputation score | No | No | No | No | No | **Yes** (Neo4j Bayesian, 0–100) |
| KYC identity verification | No | No | No | No | No | **Yes** (AES-256 identity vault) |
| Real-time WebSocket chat | Yes | Yes | Yes | Yes | Yes | **Yes** |
| Geolocation-based discovery | Yes | Yes | Yes | Yes | Yes | **Yes** (PostGIS) |
| Niyyah timer on matches | No | No | No | No | No | **Yes** (90-day nikah_year timer) |
| Family introduction milestone | No | No | No | No | No | **Yes** (`/matches/:id/family-intro`) |
| Imam directory + nikah confirmation | No | No | No | No | No | **Yes** (embedded catalog, 5 KZ cities) |
| Community social feed | No | No | No | No | No | **Yes** (posts, likes, comments) |
| Anonymous content reporting | No | Limited | Limited | No | No | **Yes** (whisper system, 3-strike) |
| Kazakh / Russian language support | No | No | No | No | No | **Yes** |
| Kazakhstan / Central Asia focus | No | No | No | No | No | **Yes** |
| Self-hosted / data sovereignty | No | No | No | No | No | **Yes** (KZ VPS, MinIO) |
| Structured courtship pathway | No | No | No | No | No | **Yes** (7-step progression) |
| Post-meeting interaction ratings | No | No | No | No | No | **Yes** (1–5, Bayesian trust update) |
| Sybil attack detection | No | No | No | No | No | **Yes** (Neo4j GDS Louvain) |

---

## 3.4 SWOT Analysis

| | **Helpful** | **Harmful** |
|---|---|---|
| **Internal** | **Strengths** | **Weaknesses** |
| **External** | **Opportunities** | **Threats** |

**Strengths.** TrueConnect's primary technical strengths are derived directly from its architectural philosophy of Islamic-first design. The mahram chat system is a genuine technical innovation: the three-way AES-256-GCM encrypted WebSocket channel is the only technically accurate implementation of mahram supervision in any dating application. The niyyah filtering algorithm enforces intention compatibility at the database query level. The Neo4j trust score engine provides a community-grounded reputation system with mathematical rigour (Bayesian smoothing, KYC-verified raters weighted at 1.5×) that no competitor offers. The self-hosted architecture guarantees data sovereignty. Clean Architecture with 102+ passing service-layer tests ensures Islamic domain business rules are verifiable and maintainable.

**Weaknesses.** The platform has not yet completed phone number OTP verification. The imam catalog and halal venue list are embedded at compile time, requiring binary redeployment to update. The trust score event channel could theoretically drop events under extreme concurrent load. The application has been tested primarily on Android; iOS testing was constrained by the absence of Apple Developer enrollment. User base is currently zero — the network effects challenge makes initial user acquisition disproportionately difficult.

**Opportunities.** Kazakhstan's Muslim matrimony market is entirely unaddressed by any locally hosted, full-featured mobile application. Growing concern about data sovereignty in Central Asia creates a tailwind for any platform that can credibly demonstrate KZ-hosted data storage. The platform's open architecture is well-suited for white-label licensing to Islamic organisations in neighbouring Uzbekistan, Kyrgyzstan, and Tajikistan.

**Threats.** Muzz, with over 10 million users, has established network effects and VC-backed marketing budgets. Match Group's resources enable rapid competitive response to emerging market segments. Cultural resistance to digital matrimony in conservative Kazakhstani communities remains a genuine adoption barrier. The single-developer origin creates a key-person risk for ongoing development and maintenance.

---

## 3.5 Gap Analysis

| Gap in Existing Systems | How TrueConnect Addresses It | Implementation in Code |
|---|---|---|
| No genuine mahram supervision channel | Three-way AES-256-GCM encrypted WebSocket channel for woman, man, and guardian | `social.mahram_chat_rooms`, `mahram_chat_msg` WS type, `internal/service/mahram_chat_service.go` |
| Niyyah field is display-only, not algorithmic | Niyyah enforced as a hard filter in the SQL discovery query | `FindCandidatesOpts.AllowedNiyyahs`, `internal/adapter/postgres/profile_repo.go` |
| No madhab-based preference matching | Same-madhab candidates receive +10 affinity score in discovery ranking | `internal/service/matching_service.go` madhab boost (B2) |
| Photo modesty not architecturally enforced | `no_photo_mode` flag triggers server-side AvatarURL suppression | `social.profiles.no_photo_mode`, `CandidateRow.NoPhotoMode`, Flutter blur render (B3) |
| No community trust / reputation system | Bayesian-smoothed Neo4j trust score (0–100), updated after each verified interaction | `internal/worker/trust_engine.go`, `internal/adapter/neo4j/trust_graph_repo.go` |
| No Sybil / fake profile detection | Neo4j GDS Louvain community detection every 6 hours; flags suspicious clusters | `internal/adapter/neo4j/trust_graph_repo.go`, `social.sybil_clusters` |
| No identity verification | KYC document upload to identity vault (AES-256 IIN, MinIO storage) | `internal/handler/kyc_handler.go`, `identity_vault.iin_vault` |
| No structured Islamic courtship pathway | Seven-step verifiable milestone progression encoded as DB flags and API endpoints | `family_intro_done`, `imam_confirmed` on `social.matches`; `/family-intro`, `/nikah-confirm` endpoints |
| No imam directory or nikah confirmation | Embedded KZ imam catalog (10 imams, 5 cities), nikah confirmation sets `married_via_app` | `internal/pkg/imam/`, `POST /v1/matches/:id/nikah-confirm` |
| No niyyah time-pressure on matches | 90-day niyyah timer on nikah_year matches, daily worker checks | `social.matches.niyyah_timer_ends_at`, `internal/worker/niyyah_timer_worker.go` |
| US/foreign data jurisdiction | All user data stored on Kazakhstani VPS; identity vault on restricted PG role | `deployments/docker-compose.yml`, `identity_vault` schema |
| No post-meeting character verification | Interaction ratings (1–5 stars, context: date/meetup/event) feed Neo4j trust score | `social.interactions`, `POST /v1/interactions`, trust engine worker |

---

## 3.6 Data Collection and User Research Methodology

The feature set of TrueConnect was shaped by a combination of religious domain knowledge, review of existing academic literature on Islamic marriage practices, and analysis of user feedback documented in the App Store and Google Play reviews of competing Islamic applications.

An informal survey of ten Muslim users in Almaty and Astana (aged 20–35, self-identified as practising Muslims seeking marriage) was conducted at the project planning stage. The results were consistent across respondents: all ten had concerns about data privacy on foreign platforms; eight of ten identified the absence of genuine mahram supervision as a blocking concern for using existing apps; all ten considered niyyah alignment important; seven of ten considered madhab compatibility a useful filter. Concerns about fake profiles and misrepresentation were raised by nine of ten respondents — directly informing the trust score and KYC feature priorities.

App Store reviews of Muzz and Salams confirmed recurring complaint patterns: the chaperone feature being "just email notifications, not real supervision"; the discovery algorithm showing clearly incompatible candidates despite niyyah profile settings; and the absence of any local imam or venue resources. These qualitative signals shaped the prioritisation of the mahram chat architecture, the niyyah filter, and the embedded imam catalog respectively.

---

## 3.7 Challenges Unique to Islamic Matrimony Applications

**Privacy Challenges.** TrueConnect resolves the tension between digital profile-based discovery and Islamic privacy norms through the `no_photo_mode` architectural feature — not as a user preference toggle but as a server-side constraint that removes the avatar URL from the discovery query result before it ever reaches the client.

**Trust Challenges.** TrueConnect's dual approach — KYC identity verification (who you are) plus community interaction ratings (what kind of person you are) — addresses both the identity dimension and the character dimension of Islamic trust verification. The Sybil detection subsystem (Neo4j GDS Louvain community detection every 6 hours) further addresses the problem of coordinated fake-profile networks.

**Cultural and Technical Challenges.** The engineering of the mahram supervision channel required solving a non-trivial technical problem: how to create a three-party authenticated communication channel over WebSocket, where each participant is independently authenticated and the Guardian's messages are visually distinguished. The solution routes `mahram_chat_msg` WebSocket messages through the same Hub infrastructure as regular chat but with a distinct message type, colour-coded in the Flutter `MahramChatScreen` (woman=green, man=blue, mahram=gold).

**Regulatory Challenges.** Data privacy legislation in Kazakhstan (the Law on Personal Data and Its Protection, 2013, amended 2023) requires that personal data of Kazakhstani citizens be stored on servers located within Kazakhstan — a requirement that TrueConnect satisfies by design through its self-hosted deployment on a Kazakhstani VPS, and that no competing platform currently meets for its Kazakhstani users.

---

# Chapter 4: Methodology

## 4.1 Development Methodology: Agile with Scrum

TrueConnect was developed using an Agile methodology structured around iterative sprints, each delivering a vertical slice of production-quality functionality. The project was organised into fourteen sprints (Sprints 0–13). Sprint 0 established the foundational scaffolding: the monorepo structure, Docker Compose orchestration, database connection pool initialisation, health check endpoints, and all inter-service connectivity. Sprint 1 delivered complete authentication infrastructure (registration, login, JWT issuance, Argon2id password hashing, AES-256-GCM PII field encryption, refresh token rotation). Sprints 2 through 5 built the core matchmaking features (profiles, discovery, matching, settings, photo upload, social feed, KYC stub, WebSocket chat, input sanitisation). Sprints 6 through 9 delivered the Islamic-domain-specific features: the halal identity schema extensions, the niyyah compatibility filter (B1), the madhab affinity boost (B2), the no-photo modesty enforcement (B3), and the niyyah 90-day timer. Sprints 10 through 11 completed the mahram chat architecture, the family introduction milestone, the embedded imam catalog, and the admin whisper-flag review system. Sprints 12 and 13 delivered the complete Flutter mobile frontend across 21 screens with the Дала Нұры design system.

User stories throughout the project were expressed in Islamic domain language: "As a Muslim woman seeking marriage, I want my photos hidden from non-mahram men until I choose to share them" — which produced the `no_photo_mode` architectural feature; "As a user with nikah_year niyyah, I want to only see candidates who share my serious marital intent" — which produced the `AllowedNiyyahs` filter; and "As a Muslim woman, I want my mahram to be an active participant in my conversations, not just notified after the fact" — which produced the mahram chat room system.

---

## 4.2 Domain-Driven Design Application

### 4.2.1 Identifying Bounded Contexts

Each TrueConnect backend module corresponds to a bounded context:

- **Auth Context** — owns identity establishment and session lifecycle; communicates cross-context only through `User.ID` UUID
- **Profile Context** — canonical owner of Islamic domain vocabulary (`niyyah`, `madhab`, `no_photo_mode`, `languages`)
- **Discovery Context** — owns the candidate recommendation algorithm; the niyyah and madhab filtering logic lives entirely within this context
- **Chat Context** — owns real-time messaging; messages are encrypted before storage and decrypted on retrieval entirely within this context
- **Mahram Context** — architecturally isolated because its access control rules are specific and cannot be generalised to the Chat context
- **Reputation Context** — owns `Interaction`, `TrustScore`, and `LeaderboardEntry`; communicates with Neo4j for graph-based score computation
- **Feed, Settings, KYC, Notifications, Whisper, Imam, and Admin Contexts** — each owning their respective domain concepts

### 4.2.2 Ubiquitous Language

The Islamic domain terms appear with identical spelling and semantics across all system layers:

**Niyyah (نية)**: `domain.NiyyahNikahYear`, `domain.NiyyahSeriousMarriage`, `domain.NiyyahFriendship` (Go enums); `niyyah` column in `social.profiles` (PostgreSQL); `AllowedNiyyahs []string` in `repository.FindCandidatesOpts`; `/niyyah` route in GoRouter; `Profile.niyyah` in Flutter model; `NiyyahBadge` widget.

**Mahram (محرم)**: `domain.Mahram` (Go struct); `social.mahrams` and `social.mahram_chat_rooms` (PostgreSQL tables); `POST /v1/mahram` API endpoint; `mahram_chat_msg` WebSocket message type; `/mahram-chat/:roomId` route; `MahramChatScreen` in Flutter.

**Madhab (مذهب)**: `domain.MadhabHanafi`, `domain.MadhabShafii`, `domain.MadhabMaliki`, `domain.MadhabHanbali`, `domain.MadhabNone` (Go enums); `madhab` column in `social.profiles`; `Profile.madhab` in Flutter; `MadhabBadge` widget.

### 4.2.3 Aggregates and Entities

- **User Aggregate**: Root `domain.User`; owns `Profile`, `UserSettings`, `RefreshToken`; deletion cascades to all owned entities
- **Match Aggregate**: Root `domain.Match`; owns `Message` entities and milestone flags (`family_intro_done`, `imam_confirmed`)
- **Post Aggregate**: Root `domain.Post`; owns `PostComment` entities and `post_likes` join table with denormalised counters
- **MahramRoom Aggregate**: Root `MahramRoom`; owns `MahramMessage` entities; three-party access control enforced by `mahram_service.go`

### 4.2.4 Repository Pattern

```
ProfileRepository     — Upsert, GetByUserID, FindCandidates, GetLeaderboard, SetMarriedViaApp
MatchRepository       — RecordLike, RecordPass, ListMatches, GetMatch, IsMatched,
                        Unmatch, BlockUser, GetBlockedIDs, GetRejectedIDs,
                        MarkFamilyIntroDone, MarkImamConfirmed, FindExpiredNiyyahMatches
MessageRepository     — Create, ListByMatch, MarkRead
PostRepository        — Create, List, GetByID, LikePost, UnlikePost, AddComment, ListComments
MahramRepository      — CreateMahram, ListMahrams, CreateRoom, ListRooms, ListMessages, SendMessage
TrustGraphRepository  — UpsertUser, RecordRating, ComputeScore, DeleteUserNode, RunSybilDetection
```

---

## 4.3 Requirements Engineering

### Functional Requirements

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-01 | Register with phone number and password | High |
| FR-02 | Issue JWT access tokens (15-min) and refresh tokens (7-day, one-time-use rotation) | High |
| FR-03 | Declare niyyah (nikah_year, serious_marriage, friendship) during onboarding | High |
| FR-04 | Filter discovery candidates to exclude niyyah-incompatible users | High |
| FR-05 | Madhab field with +10 affinity boost for same-madhab candidates | Medium |
| FR-06 | No-photo mode causing avatar suppression from all discovery results | High |
| FR-07 | Mutual-like matching via atomic PostgreSQL transaction | High |
| FR-08 | Real-time bidirectional chat via WebSocket | High |
| FR-09 | AES-256-GCM encryption of all chat messages before storage | High |
| FR-10 | Three-party supervised mahram chat room | High |
| FR-11 | Post-meeting interaction ratings (1–5 stars, context: date/meetup/event) | High |
| FR-12 | Trust score (0–100) using Bayesian-smoothed Neo4j PageRank | High |
| FR-13 | KYC identity document submission and verification | High |
| FR-14 | AES-256-GCM encrypted identity vault schema (IIN, full name) | High |
| FR-15 | Community social feed (post creation, likes, comments) | Medium |
| FR-16 | Sybil attack cluster detection using Neo4j GDS Louvain every 6 hours | High |
| FR-17 | Embedded imam directory in 5 Kazakhstani cities | Medium |
| FR-18 | Family introduction and imam nikah confirmation milestones | Medium |
| FR-19 | 90-day niyyah timer on nikah_year matches | Medium |
| FR-20 | FCM push notifications for likes and matches | Medium |
| FR-21 | Anonymous whisper reporting system with 3-strike detection | Medium |
| FR-22 | Admin endpoints for KYC verdict, Sybil review, whisper flags | High |

### Non-Functional Requirements

| ID | Requirement | Metric |
|----|-------------|--------|
| NFR-01 | API response time | < 500ms p95 |
| NFR-02 | Real-time message delivery | < 200ms end-to-end |
| NFR-03 | Authentication security | JWT HS256, Argon2id (64MB, 3 iter, 4 threads), AES-256-GCM |
| NFR-04 | Data sovereignty | All user data on Kazakhstani VPS |
| NFR-05 | Horizontal scalability | Redis pub/sub WebSocket fan-out, stateless JWT |
| NFR-06 | Rate limiting | Auth: 5 req/s; General API: 30 req/s per client |
| NFR-07 | Test coverage | 102+ service-layer tests, `go test -race` |
| NFR-08 | File upload security | MIME type validation, max 10MB |
| NFR-09 | Privacy compliance | KZ Law on Personal Data (2013, amended 2023) |
| NFR-10 | Mobile performance | 60fps swipe animations, < 1s cold start |
| NFR-11 | Offline resilience | Session restore from `flutter_secure_storage` on cold start |
| NFR-12 | Availability | > 99% uptime; Docker `restart: unless-stopped` |

---

## 4.4 Testing Strategy

All 102+ tests reside in `internal/service/` as a separate `service_test` package (black-box testing of exported service interfaces). Tests use in-memory mock implementations of all repository interfaces without requiring a live database connection. All tests are run with Go's `-race` detector enabled.

**Key test scenarios:** authentication token generation and refresh token rotation; niyyah filter exclusion (B1); no-photo mode blur flag (B3); one-sided like produces no match; mutual like produces exactly one match; self-like returns `domain.ErrInvalidInput`; block user removes match; Bayesian smoothing formula accuracy; KYC weight multiplier; Sybil cluster detection threshold; mahram third-party validation; settings default values; pagination bounds clamping.

**Manual testing:** WebSocket flows (real-time chat, mahram chat, typing indicators, read receipts) and the full Flutter UI (all 21 screens) were validated through the `make seed-halal` demo scenario (Айгерим + Алихан demo users).

---

## 4.5 Justification of Methodology

**Why Agile was necessary.** Islamic matrimony application design is a genuinely novel engineering problem. In this context, a Waterfall methodology would have produced an incorrect specification. The iterative sprint structure allowed domain requirements to emerge progressively: it was only after implementing the basic chat system in Sprint 4 that the specific technical inadequacy of a notification-based "chaperone" feature became fully clear, leading to the architectural decision to build a genuine three-party communication channel in Sprint 10.

**Why DDD was necessary.** Without DDD's concept of ubiquitous language, the Islamic terms would likely have been translated into generic software concepts, losing the semantic precision that makes the codebase understandable to both Islamic domain experts and software engineers. DDD's bounded context model provided the architectural justification for isolating the Mahram context from the Chat context.

**Why the modular monolith was preferable to microservices.** As established in the literature review [19, 20], microservices introduce distributed systems complexity only justified at a scale TrueConnect has not yet reached. The architecture is designed to support future module extraction without paying the microservices tax during the initial development phase.

---

# Chapter 5: System Architecture and Design

## 5.1 System Architecture Overview

TrueConnect is built on a three-tier architecture in which each tier has a clearly defined responsibility, a well-specified communication protocol, and a technology stack chosen to match its performance requirements.

**Tier 1 — Flutter Mobile Client** handles all user interaction, presentation logic, and client-side state management. The Flutter application communicates with the backend through two channels: HTTPS REST API calls and a persistent WebSocket connection over WSS. Authentication tokens are stored in `flutter_secure_storage` and injected automatically by the Dio HTTP client's `_AuthInterceptor`. State management is handled by Riverpod providers.

**Tier 2 — Go Backend (Modular Monolith)** contains all business logic, data access, and real-time message routing. The backend is a single deployable binary that exposes both a Gin HTTP server and a WebSocket Hub on port 8080. nginx sits in front as a reverse proxy, handling TLS termination, rate limiting, WebSocket upgrade header injection, and request routing. A separate worker binary runs background goroutines for trust score computation, Sybil detection, FCM push delivery, niyyah timer expiry, and refresh token cleanup.

**Tier 3 — Data Layer** consists of four specialised stores: PostgreSQL 16 with PostGIS for relational data and geospatial queries; Neo4j 5 Community for the trust graph and Sybil detection algorithms; Redis 7 for WebSocket pub/sub fan-out, rate limiting, seen-sets, and trust score caching; and MinIO for object storage of profile photos and KYC documents.

---

## 5.2 Backend Clean Architecture Layers

**Domain Layer** (`internal/domain/`): Pure Go structs and typed constants with zero external imports. Key types: `domain.User`, `domain.Profile` (with `Niyyah Niyyah` and `Madhab Madhab` typed fields), `domain.Match`, `domain.Message`, `domain.Post`, `domain.Interaction`, `domain.Mahram`, `domain.MahramRoom`. All error sentinels defined here: `domain.ErrNotFound`, `domain.ErrForbidden`, `domain.ErrInvalidInput`, `domain.ErrConflict`.

**Repository Layer** (`internal/repository/`): Interface definitions (ports) that the service layer depends on. `repository.ProfileRepository` with `FindCandidates(ctx, FindCandidatesOpts) ([]*CandidateRow, error)`; `repository.MatchRepository` with `RecordLike(ctx, userID, targetID uuid.UUID) (bool, uuid.UUID, error)`. The `FindCandidatesOpts` struct includes `AllowedNiyyahs []string`, making the niyyah filter a typed contract between the service and repository layers.

**Service Layer** (`internal/service/`): Business logic orchestration. Services receive repository interfaces via constructor injection and never import adapter packages. `matching_service.go` demonstrates the pattern: `GetCandidates()` calls `settingsRepo.Get()`, constructs `FindCandidatesOpts` with the niyyah compatibility list, calls `profileRepo.FindCandidates()`, then applies the madhab affinity boost (+10 to `TrustScore`) and the no-photo blur (sets `AvatarURL = ""` and `AvatarBlurred = true` when `CandidateRow.NoPhotoMode == true`) before returning `CandidateView` structs.

**Handler Layer** (`internal/handler/`): Gin HTTP handlers, middleware, and the WebSocket Hub. Handlers extract validated request parameters, call the appropriate service method, and write JSON responses. Auth middleware extracts the JWT Bearer token, verifies it, and injects `userID` into the Gin context.

**Adapter Layer** (`internal/adapter/`): Concrete implementations of repository interfaces. `adapter/postgres/` contains all SQL query implementations using `pgx/v5`. `adapter/neo4j/` implements trust graph Cypher queries. `adapter/redis/` implements cache and pub/sub. `adapter/minio/` implements the `MediaStore` interface.

---

## 5.3 Backend Module Architecture

**Auth Module** — `AuthService.Register()` hashes passwords with Argon2id (64MB, 3 iterations, 4 threads), encrypts PII with AES-256-GCM (12-byte random nonce per field), creates user records, issues JWT access tokens (HS256, 15-minute expiry), and stores Argon2id-hashed refresh tokens. Refresh token reuse detection triggers revocation of the entire token family.

**Profile Module** — `ProfileService.UpsertProfile()` validates display name and bio lengths; `ProfileService.UploadPhoto()` writes to MinIO and records the media row. Discovery query uses `ST_DWithin(p.location, ST_SetSRID(ST_MakePoint($lon, $lat), 4326)::geography, $maxDistanceMetres)`.

**Matching Module** — `MatchingService.GetCandidates()` executes a multi-step pipeline: fetch settings → load seen/rejected/blocked IDs → build `FindCandidatesOpts` with `AllowedNiyyahs` (niyyah compatibility matrix) → call `profileRepo.FindCandidates()` → apply madhab boost and no-photo blur → update seen-set cache (24h TTL).

**Chat Module** — `Hub.ServeWS()` upgrades the HTTP connection, starts a 10-second authentication timeout, waits for `{"type":"auth","token":"<JWT>"}`, registers the connection in `connections[userID]` under `sync.RWMutex`, and begins the `readPump` loop. Panic recovery at the top of each read loop prevents Hub crash from individual connection failures.

**Feed Module** — `PostService.CreatePost()` enforces the reputation gate: trust score below 30 returns 403 Forbidden. Feed listing uses cursor-based pagination on `created_at` timestamps.

**Reputation Module** — The `TrustEngine` worker drains a `chan uuid.UUID`, calls `trustGraphRepo.ComputeScore()` (Cypher query with Bayesian smoothing), and writes the result to PostgreSQL and Redis.

---

## 5.4 Database Architecture

**UUID Primary Keys** prevent enumeration attacks and enable application-layer ID generation.

**Like/Match Atomicity**: `UNIQUE(liker_id, liked_id)` on `social.likes` prevents duplicates. Mutual match detection is performed within a PostgreSQL transaction: INSERT like → check reverse like → if found, INSERT match with `UNIQUE(LEAST(user_a_id, user_b_id), GREATEST(user_a_id, user_b_id))`.

**Message Storage**: `bytea` ciphertext with 12-byte AES-GCM nonce prepended. Partial index `WHERE is_toxic = true` for admin moderation queries.

**Geolocation**: `geometry(Point, 4326)` with GIST spatial index. `ST_DWithin` with `::geography` cast for accurate great-circle distance.

**Identity Vault**: `identity_vault` schema accessible only by restricted PostgreSQL role `vault_writer`. Only `iin_hash` (Argon2id) stored unencrypted for uniqueness checks.

---

## 5.5 WebSocket Architecture

**Connection lifecycle:**
1. Client sends `GET /v1/ws` with WebSocket upgrade headers
2. nginx forwards with `Upgrade: $http_upgrade`, `proxy_read_timeout 86400s`
3. Hub validates origin against `allowedOrigins` list
4. 10-second timeout goroutine started
5. Client sends `{"type":"auth","token":"<JWT>"}`
6. `jwtManager.Verify(token)` → `connections[userID] = conn` under write lock, timeout cancelled
7. Hub sends `{"type":"auth_ok"}`
8. `readPump` goroutine begins; dispatches by `wsIncoming.Type`
9. On disconnect, connection removed from map

**Message routing**: Sender → `readPump` → service call (encrypt + persist) → direct write to recipient connection (if locally connected) → Redis PUBLISH `chat:<matchID>` (always for cross-instance delivery). Redis subscriber goroutine delivers to locally connected recipients.

**Flutter reconnection**: Exponential backoff 1s, 2s, 4s, 8s, 16s; maximum 5 attempts; then `isConnected = false` with manual reconnect button.

---

## 5.6 UML Diagram Specifications

### Diagram A: Use Case Diagram

**Actors:** Muslim User (primary), Mahram Guardian (secondary), System Administrator (secondary).

**Muslim User use cases:** Register Account; Login; Declare Niyyah (`<<include>>` Login); Edit Profile; Upload Photo; View Discovery Feed; Like Candidate; Pass Candidate; View Matches; Send Chat Message; Send Mahram Chat Message (`<<extend>>` Send Chat Message); Invite Mahram Guardian; View Community Feed; Create Post (`<<include>>` Trust Score ≥ 30); Like/Comment Post; Update Discovery Settings; Submit KYC; Rate Interaction; View Leaderboard; Report Content (Whisper); Block User; View Imam Directory; Confirm Nikah; Mark Family Introduction Done.

**System Administrator use cases:** Review KYC Submissions; Issue KYC Verdict; Review Sybil Clusters; Review Whisper Flags; List All Users.

### Diagram B: System Architecture Diagram

Three vertical columns: Flutter Client (left), Go Backend + nginx (centre), Data Stores (right). Connections: Flutter ↔ nginx via HTTPS REST + WSS; nginx → Go API port 8080 via HTTP proxy; Go API → PostgreSQL via pgx/v5 (port 5432); Go API → Neo4j via Bolt (port 7687); Go API → Redis via go-redis/v9 (port 6379); Go API → MinIO via S3 API (port 9000).

### Diagram C: Sequence — Registration and Onboarding (23 steps)

RegisterScreen → `AuthNotifier.register()` → `POST /v1/auth/register` → Argon2id hash + AES-256-GCM encrypt → INSERT `social.users` → JWT sign + refresh token INSERT → store tokens in `flutter_secure_storage` → GoRouter redirect to `/niyyah` → `PATCH /v1/profiles/me {niyyah}` → UPSERT `social.profiles` → navigate to `/home`.

### Diagram D: Sequence — Swipe Right and Match Creation (18 steps)

Swipe right → `matchingNotifier.like(candidateID)` → `POST /v1/matching/like` → `matchRepo.RecordLike()` in PostgreSQL transaction → check reverse like → if mutual: INSERT `social.matches`, send `match_notification` WS event + FCM push → Flutter shows match banner.

### Diagram E: Sequence — Real-time Chat Message (16 steps)

User A types → optimistic message appended → `channel.sink.add(chat_msg)` → Hub `readPump` → `chatSvc.SendMessage()` (AES encrypt + DB INSERT) → direct write to recipient connection OR Redis PUBLISH → Redis subscriber delivers to remote Hub instance → ChatNotifier B appends message → sender receives DB-confirmed ID replacing optimistic entry.

### Diagram F: Sequence — JWT Token Refresh (16 steps)

Dio sends request with expired token → Go returns 401 → `_AuthInterceptor.onError()` → reads refresh token from secure storage → `POST /v1/auth/refresh` → Argon2id hash verification → `MarkUsed` old token → issue new token pair → store in secure storage → retry original request with new Bearer token.

### Diagram G: Entity-Relationship Diagram

All 21 tables with FK cardinalities as defined in the database architecture section, plus Neo4j nodes `(:User)-[:RATED]->(:User)` and `(:User)-[:INTERACTED_WITH]->(:User)`.

### Diagram H: Class Diagram

Key interfaces (`ProfileRepository`, `MatchRepository`, `TrustGraphRepository`) with all method signatures; `Hub` struct with all fields and methods; `MatchingService` struct with all dependencies and methods; dependency arrows showing interface usage and implementation relationships.

---

# Chapter 6: Technology Selection and Justification

## 6.1 Backend Language and Runtime

### 6.1.1 Candidates Evaluated

Three candidates were evaluated: **Go (Golang) 1.25**, **Node.js 20 LTS (TypeScript)**, and **Python 3.12 (FastAPI/asyncio)**. Evaluation criteria were: concurrency model, memory footprint, type safety, deployment size, ecosystem maturity, and operational characteristics on a single Kazakhstan-resident VPS.

### 6.1.2 Comparative Analysis

**Table 6.1 — Backend Language Comparison**

| Criterion | Go 1.25 | Node.js 20 + TS | Python 3.12 + FastAPI |
|-----------|---------|-----------------|----------------------|
| Concurrency model | CSP goroutines (M:N scheduler, 2KB initial stack) | Event loop + Worker Threads | asyncio (single-threaded event loop) |
| Memory per idle connection | ~4 KB (goroutine stack) | ~20-40 KB (V8 heap overhead) | ~40-80 KB (Python object overhead) |
| Compiled binary size | ~8 MB static (CGO_ENABLED=0) | N/A (runtime required) | N/A (interpreter required) |
| Cold-start latency | <5 ms | ~80-120 ms (V8 JIT warm-up) | ~200-400 ms (import time) |
| Type safety | Static, checked at compile time | Optional (TypeScript transpiled) | Optional (type hints, runtime mypy) |
| Race condition detection | Built-in `-race` detector | Limited | GIL prevents true parallelism |
| Docker image size | ~12 MB (alpine + static binary) | ~180 MB (node:20-alpine + node_modules) | ~200 MB (python:3.12-slim + venv) |

**Concurrency under load.** A server with 10,000 concurrent WebSocket connections consumes roughly 20–80 MB of stack memory in Go. The equivalent Node.js implementation using a single event loop would struggle with CPU-bound tasks (trust score computations, Argon2id hashing) without offloading to Worker Threads. Python's GIL fundamentally prevents parallelism on CPU-bound work.

**Type safety and maintainability.** Go's structural typing verifies interface contracts at compile time. TypeScript offers comparable guarantees only when strict mode is enforced throughout. Python's type hints remain advisory without a CI mypy step.

**Binary deployment.** The Go backend compiles to a single static binary (~8 MB) running on `alpine:3.19` with no runtime dependency. Node.js and Python produce images typically 15–20× larger.

### 6.1.3 Decision

Go 1.25 was selected. The decisive factors were: (1) goroutine model mapping cleanly onto one-goroutine-per-connection WebSocket architecture; (2) compile-time interface enforcement enabling Clean Architecture contracts; (3) smallest deployment footprint; and (4) built-in race detection critical for the concurrent trust engine and WebSocket hub.

---

## 6.2 Mobile Framework

### 6.2.1 Candidates Evaluated

Three approaches were considered: **Flutter 3.16+ (Dart 3.2)**, **React Native 0.73 (TypeScript)**, and **native development** (Swift/Kotlin per platform).

### 6.2.2 Comparative Analysis

**Table 6.2 — Mobile Framework Comparison**

| Criterion | Flutter (Dart) | React Native (TS) | Native Swift+Kotlin |
|-----------|---------------|-------------------|---------------------|
| Rendering engine | Skia/Impeller (own canvas, pixel-perfect) | Native bridge (UIKit / Android Views) | Native UIKit / Jetpack Compose |
| UI consistency iOS/Android | Identical pixel-perfect | Near-identical (some platform inconsistencies) | Different per platform |
| Custom painter / canvas | `CustomPainter` API (full 2D canvas) | `react-native-skia` (external lib, less stable) | CoreGraphics / Canvas |
| Animation | `AnimationController`, `Tween`, built-in | `Animated` API, Reanimated 3 (external) | UIKit / Compose `animateAsState` |
| State management | Riverpod 2.x (compile-time safe, auto-dispose) | Redux Toolkit / Zustand (convention-based) | SwiftUI @State / Compose remember |
| Code sharing % | ~95% between iOS and Android | ~85% (some native modules needed) | 0% (two codebases) |
| Islamic font support | `google_fonts` (Amiri Arabic, full RTL) | `react-native-localize` + custom fonts | Built-in RTL, custom font loading |

**Custom rendering for the Дала Нұры design system.** Flutter's `CustomPainter` exposes a raw 2D canvas with direct access to `Path`, `Paint`, `Canvas.drawPath`, and transformation matrices. Three custom Islamic geometric elements were implemented entirely in Dart: `HalalPatternPainter`, `IslamicStarWidget`, and `KazakhDivider`. React Native would require `react-native-skia` (a community library with a separate release cycle); native implementations would require separate Swift and Kotlin versions, doubling maintenance cost.

**Riverpod's compile-time safety.** `AsyncNotifierProvider<MatchingNotifier, List<Map<String, dynamic>>>` is resolved at compile time; a mismatched `ref.watch()` is a compile error rather than a runtime crash. The auto-dispose lifecycle prevents memory leaks from abandoned WebSocket subscription streams.

### 6.2.3 Decision

Flutter 3.16+ with Dart 3.2 was selected. Critical factors: (1) `CustomPainter` for pixel-perfect Islamic geometric art without external dependencies; (2) Riverpod compile-time type safety for the 9-provider state graph; (3) ~95% code sharing; and (4) `google_fonts` providing the Amiri Arabic typeface with full Unicode bidirectional text support for Quranic verse display.

---

## 6.3 Primary Relational Database

### 6.3.1 Candidates Evaluated

Three databases were evaluated: **PostgreSQL 16 + PostGIS**, **MongoDB 7**, and **Firebase Firestore**.

### 6.3.2 Comparative Analysis

**Table 6.3 — Primary Database Comparison**

| Criterion | PostgreSQL 16 + PostGIS | MongoDB 7 | Firebase Firestore |
|-----------|------------------------|-----------|-------------------|
| Data model | Relational (ACID, full SQL) | Document (BSON, flexible schema) | Document (JSON, limited joins) |
| FK enforcement | Native (`FOREIGN KEY`, cascades) | Application-level only | Application-level only |
| Geospatial queries | PostGIS: `ST_DWithin`, `ST_Distance`, `GEOGRAPHY` type | `$geoNear` (2dsphere index) | Geohash-based (GeoPoint, no distance queries) |
| ACID transactions | Multi-statement, cross-table | Multi-document (4.0+, with overhead) | Limited (single document atomic) |
| Encryption at rest | `pgcrypto`, column-level AES | Field-level encryption (Enterprise) | AES-256 (managed, US jurisdiction) |
| Row-level security | Native (PostgreSQL RLS policies) | Collection-level rules | Security rules (coarser granularity) |
| Data sovereignty | Self-hosted KZ VPS | Self-hosted or Atlas | Google Cloud (US-based) |
| Two-schema isolation | Native (`social` + `identity_vault`) | Database-level separation only | No namespace isolation |

**PostGIS for proximity matching**: `ST_DWithin(p.location, ST_MakePoint($lon, $lat)::geography, $radius_m)` computes distances correctly on the spherical Earth model. Firebase's `GeoPoint` type provides no server-side distance query.

**Two-schema identity vault**: The `identity_vault` schema is accessible only to the restricted `identity_vault_role` database role. This role-based schema isolation is a native PostgreSQL feature with no equivalent in MongoDB or Firestore.

### 6.3.3 Decision

PostgreSQL 16 with PostGIS was selected. Decisive factors: (1) PostGIS `ST_DWithin` for location-based candidate filter; (2) native schema isolation for Kazakhstan data law compliance; (3) ACID multi-statement transactions for atomic like/match detection; (4) zero licensing cost with full self-hosting.

---

## 6.4 Social Graph Database

**Table 6.4 — Social Graph Options**

| Criterion | Neo4j 5 Community + GDS | PostgreSQL (recursive CTE) | Amazon Neptune |
|-----------|------------------------|---------------------------|----------------|
| Storage model | Native graph (index-free adjacency) | Adjacency list in relational tables | Property graph / RDF |
| Multi-hop traversal | O(k) pointer dereferences | O(n^k) recursive CTE | O(k) native |
| PageRank | GDS library (`gds.pageRank.stream`) | Manual iterative SQL (complex) | Gremlin `PageRankVertexProgram` |
| Louvain clustering | GDS library (`gds.louvain.stream`) | Not available natively | Not available natively |
| Data sovereignty | Self-hosted KZ VPS | Self-hosted | AWS (US jurisdiction) |
| Cost | Free (Community) | Free (included in PostgreSQL) | ~$0.10/hour minimum |

The `TrustEngine` worker calls `gds.pageRank.stream` with relationship weight set to the product of the rater's verification weight and the normalised rating value. This query runs in Neo4j's in-memory projection and returns results in milliseconds — a query that would require dozens of iterative SQL passes in PostgreSQL to converge.

**Decision:** Neo4j 5 Community with the Graph Data Science plugin. PostgreSQL cannot express PageRank or Louvain efficiently; Neptune violates Kazakhstan data sovereignty.

---

## 6.5 Real-Time Communication Protocol

**Table 6.5 — Real-Time Protocol Comparison**

| Criterion | WebSocket (RFC 6455) | SSE + HTTP POST | Long Polling |
|-----------|---------------------|-----------------|--------------|
| Connection type | Full-duplex persistent TCP | Half-duplex (server push only) | Request-response loop |
| Latency (median) | 1-5 ms | 10-30 ms (reconnect overhead) | 100-500 ms (poll interval) |
| Bidirectional | Yes | No (client POSTs separately) | No |
| Typing indicators | Trivially sent as WS message type | Requires separate POST endpoint | Not practical |
| WebRTC signaling | Directly via `webrtc_offer/answer/ice_candidate` | Possible but awkward | Not practical |
| Flutter library | `web_socket_channel` (official) | `http` package (SSE parsing) | `http` package (polling loop) |

TrueConnect's WebSocket protocol carries eight distinct client-to-server message types and seven server-to-client message types. SSE would handle server-to-client delivery adequately, but each of the eight client-to-server types would require a separate HTTP POST endpoint, introducing eight additional API surface points and doubling network round-trips for interactive operations such as typing indicators.

**Decision:** WebSocket (RFC 6455). The full-duplex requirement from WebRTC signaling alone justifies the choice; typing indicator frequency over SSE would require up to 10 POST requests per second per active typist.

---

## 6.6 Caching and Session Store

Redis 7 was selected over Memcached and PostgreSQL-based session storage based on three functional requirements:

1. **WebSocket Pub/Sub** — Redis's native Pub/Sub is used for chat fan-out. Memcached has no pub/sub primitive; PostgreSQL `LISTEN`/`NOTIFY` is limited to 8 KB payloads.
2. **Sliding-window rate limiting** — Per-user rate limiting uses Redis sorted sets with an atomic Lua script evaluating the window in a single round-trip.
3. **Candidate seen-set TTL** — Redis's native key expiration avoids a background cleanup job for the 24-hour seen-set.

**Table 6.6 — Cache Store Comparison**

| Criterion | Redis 7 | Memcached 1.6 | PostgreSQL (session table) |
|-----------|---------|---------------|---------------------------|
| Pub/Sub | Native | Not available | `LISTEN`/`NOTIFY` (8KB limit) |
| Key TTL | Per-key configurable | Per-item TTL | Requires background job |
| Sorted sets | Native (`ZADD`, `ZRANGE`) | Not available | Requires table + index |
| Lua scripting | `EVAL` atomic scripts | Not available | PL/pgSQL functions |

Redis 7 is configured with `maxmemory 256mb` and `maxmemory-policy allkeys-lru` in the Docker Compose file.

---

## 6.7 Object Storage

MinIO was selected over AWS S3 and Google Cloud Storage for a single decisive reason: **data sovereignty**. Kazakhstan's Law on Personal Data (2013, amended 2023) requires that personal data of Kazakhstan citizens be stored within the Republic of Kazakhstan. MinIO runs as a self-hosted S3-compatible object store on the same KZ-resident VPS. The S3 API compatibility means the `minio-go/v7` client SDK is API-identical to the AWS SDK, enabling a future migration without code changes if the legal landscape evolves.

---

## 6.8 Summary

**Table 6.7 — Technology Selection Summary**

| Layer | Selected | Primary Justification |
|-------|----------|-----------------------|
| Backend language | Go 1.25 | Goroutine-per-connection WS model; compile-time interface enforcement; `-race` detector |
| Mobile framework | Flutter 3.16 (Dart 3.2) | `CustomPainter` for Islamic geometric art; Riverpod compile-time safety; 95% code sharing |
| Relational database | PostgreSQL 16 + PostGIS | PostGIS proximity queries; dual-schema identity vault; ACID match atomicity |
| Graph database | Neo4j 5 + GDS | GDS PageRank and Louvain; O(k) traversal; self-hosted for KZ data sovereignty |
| Real-time protocol | WebSocket (RFC 6455) | Full-duplex for WebRTC signaling; typing indicators; Redis pub/sub fan-out |
| Cache / session store | Redis 7 | Pub/Sub; sliding-window rate limiting; seen-set TTL eviction |
| Object storage | MinIO | KZ data sovereignty for photos and KYC documents; S3-compatible API |

---

# Chapter 7: Implementation and Deployment

## 7.1 Overview

This chapter documents the concrete implementation of TrueConnect from the first migration file to the production-ready Docker Compose stack. The project was built iteratively across fourteen sprints (Sprint 0 through Sprint 13), each producing a verifiable, tested increment. The backend reached feature completion at Sprint 11 with 102 passing service-layer tests; the Flutter mobile client reached feature completion at Sprint 13 with all 21 screens implemented.

---

## 7.2 Backend Implementation

### 7.2.1 Project Bootstrapping and Database Migrations

The backend was bootstrapped as a Go module (`github.com/trueconnect/backend`) using a strict Clean Architecture directory layout. The `cmd/` directory contains three entry points — `cmd/api/main.go` (HTTP server), `cmd/worker/main.go` (background jobs), and `cmd/migrate/main.go` (schema migrations) — enabling independent deployment of each component.

Database schema evolution is managed by `golang-migrate` through 19 versioned SQL migration files. Key milestones: Migration 000001 — `social` schema with `users` table (UUID PK, Argon2id password hash, AES-GCM phone/email encryption, trust_score, verification_level enum); Migrations 000002–000008 — profiles, likes, matches, messages, posts, post_likes, comments; Migration 000009 — `identity_vault` schema with `iin_vault` and restricted PostgreSQL role; Migrations 000013–000015 — niyyah/madhab/languages/no_photo_mode columns, mahrams/mahram_chat_rooms/mahram_messages/whisper_reports tables; Migrations 000016–000019 — refresh_tokens, interactions, sybil_clusters, CASCADE deletes.

The Docker Compose service `migrate` runs as a one-shot container on every `make docker-up`, with a `depends_on: postgres: condition: service_healthy` guard preventing the API from starting before migrations complete.

### 7.2.2 Authentication Implementation

**Registration flow:** The service normalises the phone number, computes `phone_hash = Argon2id(phone, salt)` for lookup without storing plaintext, encrypts `phone_encrypted = AES-256-GCM(phone, ENCRYPTION_KEY, random_nonce)` for display recovery, hashes the password with Argon2id (64MB memory, 3 iterations, 4 parallelism threads — ~250ms per hash, raising offline dictionary attack cost to prohibitive levels), inserts the user row, and returns a JWT access token (15-minute expiry) plus a 7-day refresh token (hash stored in `social.refresh_tokens`).

**Token refresh and family revocation:** If a refresh token already marked `used_at` is presented (indicating token theft), the entire token family for that user is immediately revoked — all `social.refresh_tokens` rows for `user_id` are deleted.

**PII field encryption:** All PII fields use AES-256-GCM with a 12-byte random nonce per encryption call, ensuring two encryptions of the same plaintext produce different ciphertexts and nonce reuse is computationally infeasible.

### 7.2.3 Profile and Media Implementation

Avatar and photo uploads go through the MinIO adapter: the handler validates the MIME type against an allowlist (`image/jpeg`, `image/png`, `image/webp`), uploads to MinIO, and returns a presigned URL stored in `profiles.avatar_url`. When `no_photo_mode = true`, the profile adapter returns `AvatarURL = ""` and sets `AvatarBlurred = true` in the `CandidateView` struct, which the Flutter client renders as a blurred silhouette placeholder.

### 7.2.4 Matching Algorithm Implementation

The candidate discovery query executes five filtering stages: (1) seen-set exclusion via Redis; (2) block exclusion via `social.blocks` join; (3) geospatial filter via PostGIS `ST_DWithin`; (4) preference filters (age range, niyyah, madhab, show_me gender); (5) ordering by `trust_score DESC` with madhab affinity boost (+10 applied in Go after the SQL query).

**Niyyah timer:** When a match is created, `niyyah_timer_ends_at = NOW() + 90 days`. The `NiyyahTimerWorker` runs daily, querying matches where `niyyah_timer_ends_at < NOW()` and neither milestone is done, and sends push notification reminders. The Flutter `Match` model exposes `daysLeftOnTimer = max(0, niyyahTimerEndsAt.difference(DateTime.now()).inDays)` as a computed property, displayed as a colour-coded chip (green > 30 days, orange 8–30 days, red ≤ 7 days).

### 7.2.5 Trust Engine Implementation

When a user submits an interaction rating, the service: (1) inserts the interaction into `social.interactions`; (2) enqueues a Neo4j write via a Go channel; (3) the `TrustEngine` worker calls `gds.pageRank.stream` with interaction weight as the relationship property; (4) Bayesian-smoothes the result toward a 2.5/5 neutral baseline with `C = 5` virtual ratings; (5) scales to 0–100; (6) writes atomically to Neo4j (canonical), `social.users.trust_score` (PostgreSQL, for SQL joins), and `Redis key trust:<userId>` (sub-millisecond feed reads).

**Sybil detection** runs every 6 hours via a `time.Ticker`. Louvain community detection is projected over the full `:User`–`:INTERACTED_WITH` graph. Clusters where more than 70% of edges are internal and the KYC verification rate is below 20% are written to `social.sybil_clusters` and flagged for admin review.

### 7.2.6 Real-Time Chat Implementation

The WebSocket Hub in `internal/handler/chat_handler.go` manages one goroutine per connection. The Hub struct holds: `connections map[uuid.UUID]*websocket.Conn` (protected by `sync.RWMutex`), injected references to all services, a `*redis.Client` for pub/sub, a `*tcjwt.Manager` for token verification, and `allowedOrigins []string`.

Incoming `chat_msg` payloads are: (1) HTML-stripped via `internal/pkg/sanitize`; (2) checked by the content filter; (3) AES-256-GCM encrypted and stored in `social.messages`; (4) published to Redis channel `chat:<matchId>`. All Hub instances subscribe to `chat:*` and deliver matching messages to their locally connected users. If the recipient is not connected, the push service sends an FCM notification.

**Mahram supervision:** When a `mahram_chat_msg` is received, the Hub verifies sender participation via `mahramService.IsParticipant()`, encrypts and stores in `social.mahram_messages`, and delivers to all three room participants.

### 7.2.7 Mahram, Imam, and KYC Features

**Mahram registration** stores `mahram_phone_hash = Argon2id(phone, salt)` in `social.mahrams`. When the guardian registers from that phone number, the system resolves the hash and grants access to associated mahram rooms — avoiding plaintext phone storage while enabling lookup by phone at login time.

**Imam catalog** (`GET /v1/imams`) is served from an embedded JSON catalog compiled into the binary via `//go:embed`. The catalog contains 10 imams across 5 Kazakhstan cities. Embedding eliminates a database round-trip for every discovery screen visit.

**Nikah confirmation** (`POST /v1/matches/:id/nikah-confirm`) sets `imam_confirmed = true` and `married_via_app = true` for both users, triggering a congratulatory push notification with "Baraka Allahu feekum."

**KYC flow:** The handler validates the file MIME type (allowlist: `image/jpeg`, `image/png`, `application/pdf`), uploads to MinIO under a UUID-keyed path in the `kyc-documents` bucket (not publicly accessible), and inserts a row into `social.kyc_submissions` with status `pending`. On admin approval, `verification_level` is updated to `id_verified`, triggering the 1.5× trust weight on future interaction ratings.

---

## 7.3 Frontend Implementation

### 7.3.1 Architecture and State Management

The Flutter application uses a feature-first directory structure under `frontend/lib/features/`. All state management uses Riverpod 2.x. The `authStateProvider` is the root dependency: GoRouter's `_AuthRouterNotifier` watches it and redirects unauthenticated users to `/auth/login` on every navigation event.

### 7.3.2 HTTP Client and Interceptors

`DioClient` wraps the `Dio` HTTP client with three stacked interceptors:

1. **Auth interceptor** — reads the JWT access token from `flutter_secure_storage`, attaches it as `Authorization: Bearer <token>`. On a 401 response, calls `POST /v1/auth/refresh`, stores the new token pair, and retries the original request transparently.

2. **Retry interceptor** — retries failed requests (network errors, 5xx responses) up to three times with exponential backoff: 1s, 2s, 4s. No retry on 4xx errors.

3. **Error interceptor** — maps HTTP error codes to user-facing snackbar messages: 422 → validation error details; 429 → "Too many requests"; 5xx → "Server error, please try again."

### 7.3.3 WebSocket Client

The `chatNotifierProvider` manages the WebSocket lifecycle: opens `WebSocketChannel`, sends `{"type":"auth","token":"<JWT>"}`, subscribes to `channel.stream`, dispatches on message type. Typing indicators are debounced: repeated keystrokes within 300ms collapse to a single typing message. On connection close, exponential backoff reconnect (1s, 2s, 4s, 8s, 16s; max 5 attempts).

### 7.3.4 Key Screen Implementations

**DiscoveryScreen (`/home`):** Renders the candidate stack using a gesture detector tracking horizontal drag velocity. A filter bottom sheet exposes sliders for max distance, age range, and dropdowns for niyyah/madhab filters, wired to `settingsNotifierProvider.update()` with 450ms debounce.

**ChatScreen (`/chat/:matchId`):** `ListView.builder` with `reverse: true`; own messages in `AppColors.gold` bubbles on right; partner's in `AppColors.surface` on left. A `ContentWarningBanner` slides in on `content_warning` WS message; typing indicator shows on `isTyping = true`.

**MahramChatScreen (`/mahram-chat/:roomId`):** Three-column colour coding — woman=green, man=blue, mahram=gold. Pinned banner reads "Этот чат находится под надзором махрама" in both Kazakh and Arabic (Amiri font).

**SettingsScreen (`/settings`):** Mahram section shows guardian's obfuscated phone (last 4 digits), with Add Mahram field and delete button. All changes call `settingsNotifierProvider.update()` with 450ms debounce; on error, state is reverted to previous value.

**ImamConnectScreen (`/imams`):** City `DropdownButton` filters the imam list. Tapping "Confirm Nikah" opens a confirmation dialog; on confirmation, calls `POST /v1/matches/:id/nikah-confirm` and shows a full-screen congratulation overlay with the Quranic dua and animated 8-pointed star.

**KycScreen (`/kyc`):** Uses `image_picker` to capture/select identity document photo. Submits via `DioClient.post('/kyc/submit', formData: FormData.fromMap({'document': MultipartFile.fromFileSync(path)}))`. `kycStatusProvider` polls `GET /v1/kyc/status` every 30 seconds while in `pending` status.

### 7.3.5 Islamic Design System (Дала Нұры)

- **AppColors**: `gold = #C9A84C`, `surface = #1A1A2E`, `background = #0F0F1E`, `green = #2ECC71`, `blue = #3498DB`
- **AppTheme**: dark-first Material 3 theme with Nunito for Latin text and Amiri for Arabic/Quranic text
- **HalalPatternPainter**: tiles Қошқар мүйіз scrollwork via `Path.addArc` and `Path.cubicTo` curves
- **IslamicStarWidget**: 8-pointed star via `CustomPainter`, animated from `opacity 0.0` to `1.0` over 800ms on splash screen
- **KazakhDivider**: traditional Kazakh border motif via `Path.moveTo` and `Path.lineTo`
- **TrustScoreBadge**: `AnimatedCounter` counts from 0 to actual value over 600ms; ring colour transitions red (<40) → orange (40–70) → green (>70) via `ColorTween`

---

## 7.4 Security Implementation

**Transport Security.** nginx enforces: `X-Content-Type-Options: nosniff`; `X-Frame-Options: DENY`; `X-XSS-Protection: 1; mode=block`; `client_max_body_size 10m`; rate limiting at `30r/s` for general API and `5r/s burst=5` for authentication endpoints.

**Input Validation and Sanitisation.** All incoming request bodies are decoded into typed Go structs with `go-playground/validator/v10` tags. Validation errors map to structured 422 responses. All user-generated text is passed through `internal/pkg/sanitize.StripHTML()` before storage.

**SQL Injection Prevention.** Every database query uses parameterised queries via `pgx/v5`'s `$1`, `$2`, ... placeholder syntax. No string concatenation is used for query construction.

**CORS and Origin Validation.** CORS allowed origins are loaded from the `CORS_ORIGINS` environment variable. The WebSocket hub performs the same origin check: if `Hub.isDev = false`, connections from unlisted origins are rejected before the HTTP Upgrade.

**Content Moderation.** The content filter checks both Arabic-script and Latin-script patterns. Messages classified as `is_toxic = true` are stored with the flag set; a `content_blocked` WebSocket event is sent to the sender. The recipient never receives the message.

---

## 7.5 Deployment Pipeline

### 7.5.1 Docker Compose Stack

| Service | Image | Port | Notes |
|---------|-------|------|-------|
| postgres | `postgis/postgis:16-3.4-alpine` | 5433:5432 | PostGIS pre-installed |
| neo4j | `neo4j:5-community` | 7474, 7687 | GDS plugin mounted via volume |
| redis | `redis:7-alpine` | 6379 | `maxmemory 256mb`, `allkeys-lru` |
| minio | `minio/minio:latest` | 9000, 9001 | S3-compatible object storage |
| migrate | Custom (`cmd/migrate`) | — | One-shot; `depends_on postgres healthy` |
| api | Custom (`cmd/api`) | 8080 | `depends_on migrate` |
| nginx | `nginx:alpine` | 80:80 | Reverse proxy + rate limiting |

The API Dockerfile uses a two-stage build: builder stage (`golang:1.25-alpine`) compiles with `CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w"`; runtime stage (`alpine:3.19`) copies only the binary, running as non-root user `appuser` (UID 1000).

### 7.5.2 Nginx Configuration

Rate limiting zones:
```nginx
limit_req_zone $binary_remote_addr zone=api_limit:10m  rate=30r/s;
limit_req_zone $binary_remote_addr zone=auth_limit:10m rate=5r/s;
```

WebSocket proxy with 86400-second timeouts prevents nginx from terminating idle connections between chat sessions:
```nginx
location /v1/ws {
    proxy_pass http://api_backend;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_read_timeout 86400s;
    proxy_send_timeout 86400s;
}
```

### 7.5.3 Continuous Integration

The CI pipeline (`.github/workflows/ci.yml`) runs on every push and pull request to `main`:
- `go build ./cmd/api/...` and `go build ./cmd/worker/...` — verify compilability
- `go test ./internal/... -race -count=1` — run all 102+ tests with race detector
- `go vet ./...` — static analysis

The `-count=1` flag disables test result caching, ensuring every CI run executes all tests.

### 7.5.4 Makefile Developer Experience

| Target | Purpose |
|--------|---------|
| `build` | Compile all three binaries |
| `run` | Start API server locally |
| `run-worker` | Start background worker |
| `test` | Run all tests with `-race` |
| `lint` | Run `golangci-lint` |
| `migrate-up` / `migrate-down` | Apply / rollback migrations |
| `docker-up` / `docker-down` | Start / stop full Docker stack |
| `docker-reset` | Full reset including volumes |
| `docker-logs` | Tail all service logs |
| `seed-halal` | Load demo users (Айгерим + Алихан) |
| `reset-demo` | Clean demo environment end-to-end |

The `seed-halal` target loads `migrations/seed_halal_demo.sql` containing two Almaty users with a pre-created match, a mahram room, and sample messages.

---

## 7.6 Testing Summary

**Table 7.1 — Test Coverage by Sprint**

| Sprint | Feature Tested | New Tests | Cumulative |
|--------|---------------|-----------|------------|
| 1 | Auth (register, login, refresh, logout, brute-force) | 30 | 30 |
| 2 | Profiles CRUD, PostGIS candidates, like/pass, settings | 21 | 51 |
| 3 | Interactions, trust score, Sybil detection, worker | 19 | 70 |
| 4 | WebSocket chat, feed CRUD, KYC stub, message encryption | 25 | 95 |
| 5 | Token family revocation, rate limiting, reports | 4 | 99 |
| 6–8 | Mahram, whisper, settings adapters | 19 | 118 |
| 9–11 | Halal filters, mahram chat, family intro, imam catalog | 7 | 120+ |

All 102 active service-layer tests pass with `-race -count=1`. The test suite uses mock repository implementations satisfying all repository interfaces, enabling `MatchingService`, `AuthService`, and `ReputationService` tests to run without a live database connection.

---

## 7.7 Demo Walkthrough

The complete end-to-end demo scenario, reproducible via `make docker-up && make seed-halal`:

1. **Launch** — `flutter run -d android` opens the splash screen with the animated 8-pointed Islamic star and the Quranic verse Ar-Rum 30:21.
2. **Onboarding** — three slides with Қошқар мүйіз CustomPainter background; tap "Начать" to navigate to NiyyahSelection.
3. **Login** — enter Айгерим's demo credentials; `authStateProvider` calls `POST /v1/auth/login`, stores tokens in `flutter_secure_storage`, routes to `/home`.
4. **Discovery** — Алихан's card appears at the top of the stack (PostGIS candidate query within 50 km, same hanafi madhab → +10 boost). Swipe right triggers `POST /v1/matching/like`; the server detects a mutual like, creates a match, sends `match_notification` WS event.
5. **Match banner** — a gold congratulation overlay slides up with both users' avatars.
6. **Chat** — navigate to `/chat/<matchId>`; type a message → real-time delivery via WebSocket.
7. **Mahram invite** — tap the guardian invite button; enter the guardian's phone number → `POST /v1/mahram`; guardian joins; chat transitions to 3-way MahramChatScreen.
8. **Settings** — navigate to `/settings`; adjust age range; observe the discovery feed update on return.
9. **KYC** — navigate to `/kyc`; upload a document photo; status shows "Pending" then transitions to "Approved" after admin verdict.
10. **Imam Connect** — navigate to `/imams?matchId=<id>`; select Almaty; tap imam card → confirm nikah; "Baraka Allahu feekum" success overlay with animated star.

This scenario exercises all 15 backend modules, all 50 API endpoints, the WebSocket hub, Redis pub/sub, Neo4j trust score, PostGIS candidate query, MinIO upload, and all 21 Flutter screens.

---

# References

[1] E.J. Finkel, P.W. Eastwick, B.R. Karney, H.T. Reis, and S. Sprecher, "Online Dating: A Critical Analysis from the Perspective of Psychological Science," *Psychological Science in the Public Interest*, vol. 13, no. 1, pp. 3–66, 2012.

[2] G. Tyson, V.C. Perta, H. Haddadi, and M.C. Seto, "A First Look at User Activity on Tinder," in *Proc. IEEE/ACM International Conference on Advances in Social Networks Analysis and Mining (ASONAM)*, pp. 461–466, 2016.

[3] G.J. Hitsch, A. Hortaçsu, and D. Ariely, "Matching and Sorting in Online Dating," *American Economic Review*, vol. 100, no. 1, pp. 130–163, 2010.

[4] N. Ellison, R. Heino, and J. Gibbs, "Managing Impressions Online: Self-Presentation Processes in the Online Dating Environment," *Journal of Computer-Mediated Communication*, vol. 11, no. 2, pp. 415–441, 2006.

[5] C.L. Toma, J.T. Hancock, and N.B. Ellison, "Separating Fact From Fiction: An Examination of Deceptive Self-Presentation in Online Dating Profiles," *Personality and Social Psychology Bulletin*, vol. 34, no. 8, pp. 1023–1036, 2008.

[6] Pew Research Center, "The Future of World Religions: Population Growth Projections, 2010–2050," Washington D.C.: Pew Research Center, 2015.

[7] G.R. Bunt, *iMuslims: Rewiring the House of Islam*, Chapel Hill: University of North Carolina Press, 2009.

[8] E. Kavakci and C.R. Kraeplin, "Religious Beings in Fashionable Spaces: The Online Identity of Salafi Women in the United States," *New Media & Society*, vol. 19, no. 9, pp. 1441–1458, 2017.

[9] Y. Al-Saggaf, "Online Community Standards and Ethics: An Islamic Perspective," in *Digital Islam*, E. Baulch, Ed., New York: Routledge, 2019.

[10] J. Bobadilla, F. Ortega, A. Hernando, and A. Gutiérrez, "Recommender Systems Survey," *Knowledge-Based Systems*, vol. 46, pp. 109–132, 2013.

[11] D. Gale and L.S. Shapley, "College Admissions and the Stability of Marriage," *American Mathematical Monthly*, vol. 69, no. 1, pp. 9–15, 1962.

[12] R. Burke, "Hybrid Recommender Systems: Survey and Experiments," *User Modeling and User-Adapted Interaction*, vol. 12, no. 4, pp. 331–370, 2002.

[13] M. Ye, P. Yin, W.-C. Lee, and D.-L. Lee, "Exploiting Geographical Influence for Collaborative Point-of-Interest Recommendation," in *Proc. ACM SIGIR*, pp. 325–334, 2011.

[14] I. Fette and A. Melnikov, "The WebSocket Protocol," IETF RFC 6455, Dec. 2011.

[15] R.T. Fielding and R.N. Taylor, "Principled Design of the Modern Web Architecture," *ACM Transactions on Internet Technology*, vol. 2, no. 2, pp. 115–150, May 2002.

[16] J.L. Carlson, *Redis in Action*, Shelter Island: Manning Publications, 2013.

[17] R.C. Martin, *Clean Architecture: A Craftsman's Guide to Software Structure and Design*, Upper Saddle River: Prentice Hall, 2017.

[18] E. Evans, *Domain-Driven Design: Tackling Complexity in the Heart of Software*, Boston: Addison-Wesley, 2003.

[19] S. Newman, *Building Microservices*, 2nd ed., Sebastopol: O'Reilly Media, 2021.

[20] M. Fowler and J. Lewis, "Microservices," *martinfowler.com*, Mar. 2014. [Online]. Available: https://martinfowler.com/articles/microservices.html

[21] A. Biørn-Hansen, T.A. Majchrzak, and T.-M. Grønli, "Progressive Web Apps vs. Native Mobile Apps: A Multi-Criteria Comparison," in *Proc. 14th International Conference on Mobile Web and Intelligent Information Systems (MobiWis)*, pp. 89–98, 2017.

[22] Google, "Flutter: Build Apps for Any Screen," *flutter.dev*, 2018. [Online]. Available: https://flutter.dev

[23] P. Nawrocki, K. Wrona, M. Marczak, and P. Szmeja, "Comparison of Native and Cross-Platform Frameworks for Mobile Applications," in *Proc. 16th International Conference on Software Engineering Advances (ICSEA)*, pp. 96–106, 2021.
