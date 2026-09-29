# 100 Scenario Stress Test

## A. Inbound (1--15)

1.  normale Plaintext-Mail
2.  normale HTML-Mail
3.  Multipart alternative
4.  Inline-CID-Bild
5.  20 Attachments
6.  sehr großer Anhang
7.  kaputtes MIME
8.  fehlende Message-ID
9.  doppelte Message-ID
10. gleiche Mail zweimal zugestellt
11. Edge Event doppelt
12. Parser stirbt mitten im Job
13. DB fällt vor Commit aus
14. Object Store fällt während Attachment Commit aus
15. unbekannter Empfänger

## B. Outbound (16--30)

16. normaler Send
17. SES temporär nicht erreichbar
18. Timeout nach Providerannahme
19. permanenter Recipient Reject
20. Bounce nach Annahme
21. Complaint
22. Doppelklick Send
23. Client retryt Send
24. Worker stirbt während Send
25. Scheduled Send Client aus
26. Scheduled Send Core restart
27. Scheduled Send ändern
28. Scheduled Send im letzten Moment abbrechen
29. 100 Empfänger
30. Attachment Vault Link widerrufen

## C. Sync / Devices (31--45)

31. Mac offline 1 Stunde
32. Smartphone offline 3 Monate
33. Gerät verliert lokalen Cache
34. Gerät wird widerrufen
35. zwei Geräte lesen gleichzeitig
36. read/unread Konflikt
37. Label add/remove Konflikt
38. Delete während anderes Gerät offline
39. alter Cursor
40. Event doppelt
41. Event verspätet
42. Event Relay komplett weg
43. Rebootstrap
44. Snapshot + Delta
45. 5 Geräte gleichzeitig

## D. Drafts (46--55)

46. Draft Mac -\> Smartphone
47. Smartphone offline editiert
48. zwei Geräte gleiche Revision
49. Attachment Upload während Draft
50. Upload schlägt fehl
51. Draft wird gesendet während anderes Gerät offen
52. Autosave verliert Netz
53. 1-MB-Markdown-Draft
54. Draft löschen auf Gerät A
55. Konfliktfassung wiederherstellen

## E. Attachments (56--65)

56. Vault offline
57. Datei bereits dedupliziert
58. Hash mismatch
59. Malwareverdacht
60. Inline Content remote
61. Cache voll
62. Offline Pin
63. referenziertes Objekt purge-kandidat
64. EasyPeasyDrop abgelaufen
65. Capability Token geleakt/widerrufen

## F. Identity / Security (66--75)

66. falsches Passwort mehrfach
67. gestohlene Session
68. Device revoke
69. Recovery Code
70. zweiter Passkey
71. verdächtiger neuer Login
72. Mass Export
73. Weiterleitung heimlich aktiviert
74. Rate-limit Angriff
75. kompromittierter Account beginnt Spam

## G. Infrastructure (76--90)

76. API Container tot
77. Worker Container tot
78. Web Container tot
79. PostgreSQL restart
80. DB korrupt
81. Object Store down
82. SES outbound down
83. SES inbound down
84. DNS fehlerhaft
85. Zertifikat läuft ab
86. voller Datenträger
87. Queue Flood
88. Region-Ausfall
89. Backup fehlt
90. Restore auf leerer Infra

## H. Product / Attention (91--100)

91. 100.000 Newsletter
92. Newsletter falsch klassifiziert
93. persönliche Mail als Newsletter
94. Spam als persönlich
95. Conscious mit technischer Störung
96. Development Debug View
97. BurnOut mit voller Timeline
98. Print Conscious
99. `.eml` Export + Reimport
100. Wechsel des Moodivadors verändert keinerlei kanonischen State

## Bestehensregel

Jedes Szenario bekommt später: - Expected Result - Datenverlust:
ja/nein - Nutzerwirkung - automatische Recovery - manueller Eingriff -
Testautomatisierung - gemessene RTO
