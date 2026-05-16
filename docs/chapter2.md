# Chapter 2: Literature Review

## 2.1 Online Dating Platforms: Evolution and Behavioural Research

The academic study of online dating has matured considerably since the first wave of empirical research in the mid-2000s, evolving from descriptive accounts of early platform adoption to sophisticated analyses of algorithmic influence on mate selection, user self-presentation strategies, and the psychological consequences of gamified matchmaking. Understanding this literature is essential context for TrueConnect, because the platform is in many respects a deliberate architectural counter-argument to the design patterns that empirical research has identified as dominant in the industry.

The foundational comprehensive review of online dating science was conducted by Finkel, Eastwick, Karney, Reis, and Sprecher [1], whose analysis in Psychological Science in the Public Interest remains the most cited critical treatment of the field. The authors identify three categories of claims made by online dating platforms: that computer-mediated communication is superior to face-to-face communication for initial partner evaluation; that access to a large pool of potential partners improves matching outcomes; and that mathematical compatibility algorithms improve matching accuracy. Their systematic review finds the empirical support for each of these claims to be weaker than platform marketing suggests. Crucially for the TrueConnect design, Finkel et al. note that algorithmic matching systems tend to reduce potential partners to sets of profile attributes and thereby fail to capture the interactive chemistry that predicts long-term relationship satisfaction. This finding reinforces the TrueConnect design decision to use constraint-based niyyah and madhab filtering not as a similarity optimisation but as a requirement filter — ensuring that users are never shown candidates with fundamentally incompatible intentions — while leaving the qualitative dimensions of compatibility to emerge through supervised conversation.

The analysis of actual user behaviour on contemporary swipe-based platforms by Tyson, Perta, Haddadi, and Seto [2] provides quantitative evidence of the pathological dynamics that gamified discovery mechanics create. Analysing Tinder data across multiple metropolitan areas, they find extreme asymmetry in like rates between male and female users and demonstrate that the swipe mechanic creates behaviour more closely resembling slot-machine engagement than deliberate partner evaluation. Male users swipe right on approximately 46 percent of profiles, while female users approve approximately 14 percent — a disparity that the authors argue is a direct consequence of the interface design rather than underlying preference differences. TrueConnect's discovery screen (`DiscoveryScreen` using `FlutterCardSwiper`) intentionally preserves the familiar swipe metaphor while removing the like-volume incentive: there are no like counters, no match-percentage badges, and no "super like" mechanics that commoditise interest expression.

The economics of online partner matching were rigorously analysed by Hitsch, Hortaçsu, and Ariely [3] in a study using data from a major US online dating platform. Their analysis demonstrates that revealed preferences in partner selection are strongly driven by income, physical attractiveness (as measured by profile photo ratings), and height for male candidates — attributes that have limited relevance in Islamic spouse selection where piety, family background, madhab compatibility, and declared intention are given priority. This divergence between the attribute dimensions rewarded by mainstream matching optimisation and the attribute dimensions prioritised in Islamic courtship is a central argument for building a domain-specific platform rather than customising an existing one.

Self-presentation dynamics in online dating profiles have been extensively studied. Ellison, Heino, and Gibbs [4] find that users engage in a process of "selective self-presentation" — revealing information strategically to manage impressions while attempting to maintain plausibility for eventual face-to-face encounters. This creates a tension between optimistic self-presentation and the Islamic prohibition on deception (ghish). Toma, Hancock, and Ellison [5] extend this analysis with objective measurements, finding that male users systematically overstate height and income while female users understate weight in their profiles. TrueConnect's KYC identity verification system (`social.kyc_submissions`, `identity_vault.iin_vault`) directly addresses this deception dynamic by providing a verifiable identity layer: a KYC-verified badge (communicated via `verification_level` in `social.users`) signals to other users that the profile has been cross-checked against a government identity document, establishing a baseline of factual accuracy that unverified platforms cannot offer.

---

## 2.2 Islamic Marriage Practices and the Role of Technology

Islamic marriage is governed by a rich and well-documented jurisprudential tradition that prescribes specific roles, processes, and ethical constraints absent from secular matchmaking. Engagement with this literature is necessary to justify the specific domain design decisions made in TrueConnect — decisions that might appear unusual to a secular software engineer but are precisely motivated by religious requirements.

The global Muslim population context is provided by the Pew Research Center's comprehensive demographic analysis [6], which projects that the Muslim population will reach 2.76 billion by 2050, representing 29.7 percent of the global total, with particularly rapid growth in Sub-Saharan Africa and sustained majority status in Central Asia. Kazakhstan is specifically identified as a Muslim-majority country (approximately 70 percent of the population identifying as Muslim), reinforcing the market rationale for a locally developed platform. The report also provides data on the younger age profile of Muslim populations globally — a demographic that is simultaneously the primary target user of matrimony applications and the cohort most comfortable with digital-first relationship formation.

The broader relationship between Islam and digital technology has been studied by Bunt [7], whose work on what he terms "iMuslims" documents the extensive and sophisticated use of digital networks by Muslim communities for religious learning, community organisation, and social connection. Bunt argues that digital Islam is not a dilution of traditional practice but an extension of the Islamic tradition of adapting available technology to religious purposes. This framing legitimises the TrueConnect project as a continuation of a documented pattern of Islamic digital practice rather than a novelty or cultural compromise.

The specific challenges of identity and religious performance in online spaces for Muslim women have been studied by Kavakci and Kraeplin [8], who document the strategies used by observant Muslim women to maintain religious identity in digital social environments that were not designed with their needs in mind. Their work directly informs the TrueConnect `no_photo_mode` feature: rather than designing a platform that forces observant women to choose between visibility and modesty, TrueConnect treats modesty control as an architectural requirement. When `no_photo_mode = true` in `social.profiles`, the discovery query returns `AvatarURL = ""` and sets `NoPhotoMode = true` in the `CandidateRow` struct, which the Flutter `DiscoveryScreen` renders as a blurred placeholder — technically equivalent to an observant woman choosing to introduce herself by character before appearance.

The ethics of digital communication in Islamic contexts have been addressed by Al-Saggaf [9], who argues that Islamic online communities require specific ethical infrastructure — mechanisms for accountability, truthfulness, and guardianship — that secular platforms do not provide. This argument directly maps to the three technical innovations in TrueConnect that have no parallel in secular dating applications: the mahram chat room system (guardian accountability), the KYC identity verification (truthfulness infrastructure), and the trust score engine (community-based reputation). The correspondence between the ethical requirements identified in the academic literature and the technical features implemented in the codebase provides strong post-hoc validation of the design decisions made.

---

## 2.3 Recommendation and Matching Algorithms

The algorithmic core of any matchmaking platform is its candidate recommendation and ranking system. A thorough engagement with the recommender systems literature is necessary both to situate TrueConnect's matching approach and to explain the deliberate choice not to use collaborative filtering — the dominant paradigm in commercial recommendation systems — in the Islamic matrimony context.

The most comprehensive survey of recommender system approaches is provided by Bobadilla, Ortega, Hernando, and Gutiérrez [10], who classify recommender systems into three main paradigms: collaborative filtering (predicting user preferences from the preferences of similar users), content-based filtering (recommending items similar to those a user has already approved), and hybrid systems that combine both. Collaborative filtering has well-documented strengths in domains like music or e-commerce recommendation, where the cold-start problem is its primary weakness. In the matrimony domain, however, collaborative filtering would mean surfacing candidates whom similar users have liked — a mechanism that is both ethically problematic (implying that your romantic preferences are predictable from a demographic cluster) and practically counterproductive for the Islamic use case where individual religious compliance, family background, and declared intention are not capturable in aggregate similarity metrics.

TrueConnect's matching algorithm implements a form of constraint-based content filtering: the `FindCandidates` query in `internal/adapter/postgres/profile_repo.go` first applies hard constraint filters (niyyah compatibility via the `AllowedNiyyahs` list, bidirectional block exclusion, seen-set exclusion, geographic distance via PostGIS `ST_DWithin`) and then applies a soft preference boost (madhab affinity +10 to `TrustScore`) to rank candidates within the constraint-satisfying set. This is analogous to the preference-based filtering approach described in the hybrid recommender literature but grounded in domain-specific Islamic constraints rather than learned behavioural patterns.

The theoretical foundation for stable two-sided matching was established by Gale and Shapley [11] in their seminal paper on college admissions, which introduced the concept of a stable matching — a pairing where no two unmatched individuals both prefer each other to their current assignment. The Gale-Shapley algorithm is the theoretical basis for the "matching market" view of dating platforms. TrueConnect's like/pass mechanism can be understood as implementing a version of this model: two users must mutually express interest (the `RecordLike` function returning `matched=true` when the second like is recorded) before a match is created, ensuring that no match exists where one party is paired with someone they have explicitly rejected. The niyyah timer system (`niyyah_timer_ends_at` on `social.matches`) further refines this by introducing a temporal constraint that dissolves matches between users who do not progress to a family introduction within the declared timeline.

Hybrid recommender systems have been surveyed by Burke [12], who identifies seven distinct hybridisation strategies for combining content-based and collaborative approaches. The weighted hybrid — assigning scores from multiple recommendation techniques and combining them linearly — most closely describes TrueConnect's graph-candidate system (`GET /v1/matching/graph-candidates`), which uses Neo4j PageRank-weighted trust scores to surface candidates who are highly regarded by the community while satisfying content-based filters.

Location-based social network (LBSN) recommendation is directly relevant to TrueConnect's geolocation-based discovery. Ye, Yin, Lee, and Lee [13] demonstrate that incorporating geographic influence into collaborative filtering significantly improves recommendation accuracy for location-sensitive social applications. TrueConnect's PostGIS-based distance query (filtering by `max_distance_km` from `social.user_settings`) implements the geographic constraint dimension of this research, ensuring that candidates are physically reachable for the in-person meetings that the Islamic courtship process requires.

---

## 2.4 Real-Time Communication Architecture

The real-time communication subsystem of TrueConnect — encompassing the WebSocket Hub, Redis pub/sub fan-out, and AES-256-GCM message encryption — represents a significant engineering investment. The design choices made in this subsystem are grounded in a well-established body of literature on real-time web communication architectures.

The WebSocket protocol, defined in IETF RFC 6455 by Fette and Melnikov [14], provides a full-duplex communication channel over a single TCP connection, fundamentally distinguishing it from the HTTP request-response paradigm. The protocol performs an HTTP/1.1 upgrade handshake and then transitions to a framing protocol that allows either party to send data at any time without the overhead of repeated HTTP headers. For TrueConnect's chat system, WebSocket is the only technically appropriate choice: the mahram chat requirement mandates that a message sent by any of three participants (woman, man, guardian) is delivered to all three simultaneously, which requires genuine server-push capability that HTTP polling and Server-Sent Events cannot provide without significant complexity and latency overhead.

The REST architectural style, described by Fielding and Taylor [15], is used for all non-real-time operations in TrueConnect (profile management, match creation, feed, notifications). The clean separation between REST API operations (using Gin HTTP handlers in `internal/handler/`) and WebSocket operations (using the Hub in `internal/handler/chat_handler.go`) reflects the principle articulated by Fielding that architectural styles should be matched to their appropriate communication patterns rather than applied uniformly across all system interactions.

Redis as a messaging backbone for real-time systems is described in practical detail by Carlson [16]. TrueConnect uses Redis pub/sub with a per-match channel naming scheme (`chat:<matchID>`) to enable horizontal scaling: when a message is sent via WebSocket from User A to User B, the Hub first attempts direct delivery to User B's local connection. If User B is connected to a different server instance, the message is PUBLISHED to the Redis channel, which is SUBSCRIBED by all Hub instances, allowing the correct instance to deliver the message. This architecture is directly implemented in `internal/adapter/redis/` and the Hub's Redis subscriber goroutine. The same pattern is used for the mahram chat channel, with the `mahram_chat_msg` WebSocket message type routed through the same Redis pub/sub infrastructure to all three participants.

Mobile-specific considerations for real-time communication — battery consumption, network switching, and background connectivity — are not extensively covered by the Web-centric literature but are addressed in the TrueConnect Flutter implementation through the exponential backoff reconnection strategy implemented in `chat_provider.dart` (delays: 1s, 2s, 4s, 8s, 16s, maximum 5 attempts). This prevents battery drain from rapid reconnection attempts while ensuring that a temporary network interruption (such as transitioning between WiFi and mobile data) does not permanently break the chat connection.

---

## 2.5 Software Architecture for Complex Domains

The architectural decisions that shape TrueConnect — Clean Architecture, Domain-Driven Design, and a modular monolith deployment pattern — are grounded in a mature body of software engineering literature. Understanding these foundations is essential for evaluating the design choices made throughout the system.

Clean Architecture, as articulated by Robert C. Martin [17], prescribes a concentric layer model in which the innermost layer contains pure business entities with no framework dependencies, and each outer layer depends only on inner layers — never the reverse. In TrueConnect, this manifests as the dependency rule: `internal/domain/` has zero external imports; `internal/service/` depends only on `internal/domain/` and `internal/repository/` interfaces; `internal/handler/` depends on services; and `internal/adapter/` implements repository interfaces but is never imported by services. This architecture provides a critical benefit for a domain as semantically rich as Islamic matrimony: the business rules (niyyah compatibility, mahram third-party validation, trust score Bayesian smoothing) are isolated in the service layer and can be tested independently of database or HTTP framework concerns, as demonstrated by the 102+ passing service-layer tests.

Domain-Driven Design (DDD), introduced by Evans [18], provides the vocabulary and structuring principles that allow a complex domain to be faithfully reflected in code. The central DDD concept of ubiquitous language — using the same terms in code, database schema, API design, and team communication — is visibly implemented throughout TrueConnect. The Islamic domain terms `niyyah`, `mahram`, and `madhab` appear identically in Go struct field names (`domain.Profile.Niyyah`), PostgreSQL column names (`social.profiles.niyyah`), API endpoint paths (`/v1/mahram`, `/v1/matching/likes`), Flutter model fields (`Profile.niyyah`), and screen route names (`/niyyah`). This terminological consistency, which Evans argues is a precondition for a domain model that accurately reflects business requirements, is only achievable when the development team treats domain language as a first-class engineering artefact rather than a naming convention.

The modular monolith pattern — a single deployable unit internally organised as loosely coupled modules — occupies an important position in the microservices literature as the appropriate intermediate architecture for systems that need clear module boundaries but do not yet have the operational complexity or team size to justify microservice deployment. Newman [19] argues that a modular monolith is frequently the correct architectural choice for a system in its early growth phase: it preserves the option to extract modules into independent services later while avoiding the distributed systems complexity (network partitions, service discovery, distributed tracing) that microservices introduce from day one. TrueConnect's modular monolith — with 15 distinct modules (`auth`, `profile`, `matching`, `chat`, `feed`, `settings`, `mahram`, `reputation`, `kyc`, `notifications`, `whisper`, `imam`, `admin`, `user`) as Go packages within a single binary — exemplifies this principle.

Fowler and Lewis [20] provide the canonical definition of microservices and the criteria by which an organisation should consider transitioning from a monolith. Their "micro" criteria — independently deployable, organised around business capabilities, built by small teams — would apply to TrueConnect at a future scale (millions of users, multiple development squads) but are not met by the current project. The modular monolith correctly anticipates this eventual transition: each of the 15 TrueConnect modules already has a well-defined interface boundary (its service constructor and exported methods), which would constitute the contract for a future independent service deployment.

---

## 2.6 Cross-Platform Mobile Development

The choice to build TrueConnect's mobile client with Flutter (rather than native Kotlin/Android or a React Native cross-platform approach) was made on technical grounds that are well-supported by the comparative mobile development literature.

Biørn-Hansen, Majchrzak, and Grønli [21] provide a systematic comparison of progressive web applications and native mobile applications across multiple criteria, concluding that the choice of mobile development approach should be driven by the performance and user experience requirements of the specific application. Their framework, applied to TrueConnect, yields a clear recommendation toward compiled native or near-native approaches: the swipe-based discovery interface (`DiscoveryScreen` using `FlutterCardSwiper`) requires smooth 60fps card animations; the mahram chat screen (`MahramChatScreen`) requires real-time message delivery with sub-200ms visual update; and the trust score badge (`TrustScoreBadge` with `AnimatedCounter`) requires smooth numerical animation.

Flutter, as documented in the technical literature by Google [22], achieves near-native performance through a key architectural distinction: unlike React Native, which renders using the platform's native UI components via a JavaScript bridge, Flutter uses its own rendering engine (Skia, migrating to Impeller for GPU-accelerated rendering) that compiles directly to native ARM code. This means that Flutter animations run in the Dart VM with direct access to the GPU, without the latency overhead of a JavaScript bridge. For TrueConnect's animation-heavy swipe interface, this distinction is not merely theoretical — it is the reason why the card swiper remains smooth at 60fps even when the matching state is being updated by Riverpod providers in the same frame.

Riverpod, the state management solution used throughout TrueConnect's Flutter frontend, was chosen over the alternatives (Provider, BLoC, GetX) for reasons articulated in the Flutter community literature. Nawrocki, Wrona, Marczak, and Szmeja [23] compare the architectural characteristics of cross-platform frameworks and find that type-safe compile-time dependency injection — which Riverpod provides through its `Provider` declarations and `ConsumerWidget` binding — is particularly valuable for applications with complex interdependent state, such as an app where authentication state (`authStateProvider`) must gate access to discovery state (`matchingNotifierProvider`), chat state (`chatNotifierProvider`), and settings state (`settingsNotifierProvider`) simultaneously. Flutter's single-codebase model also directly supports the future iOS release of TrueConnect at minimal additional development cost.

---

## 2.7 Literature Review Summary

The literature surveyed in this chapter converges on several conclusions that directly validate the design decisions made in TrueConnect.

Empirical research on online dating behaviour [1, 2, 3] establishes that existing platforms' gamification mechanics and photograph-first design produce outcomes that are misaligned with the needs of users seeking serious, values-aligned partners. The Islamic marriage scholarship [6, 7, 8, 9] confirms that Muslim users require specific technical infrastructure — guardian supervision, identity verification, deception prevention — that no existing platform provides. The recommender systems literature [10, 11, 12, 13] justifies TrueConnect's constraint-based filtering approach over collaborative filtering for the Islamic domain. The real-time communication literature [14, 15, 16] supports the WebSocket + Redis pub/sub architecture as the correct technical choice for the mahram three-party supervised chat requirement. The software architecture literature [17, 18, 19, 20] validates the Clean Architecture + DDD + modular monolith pattern as appropriate for a system of this complexity and scale. The mobile development literature [21, 22, 23] confirms Flutter as the correct platform choice given TrueConnect's animation and state management requirements.

The specific gap that this project addresses — and that the surveyed literature does not solve — is the absence of a production-quality, self-hosted mobile platform that treats Islamic domain requirements as first-class architectural constraints rather than optional features. TrueConnect fills this gap through the integration of technical choices drawn from each of the literature areas surveyed, unified by a domain model that places the Islamic courtship process at the centre of every design decision.

---

## References

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
