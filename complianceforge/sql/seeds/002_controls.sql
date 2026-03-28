-- ============================================================
-- ComplianceForge — Seed: ISO/IEC 27001:2022 Annex A Controls
-- All 93 controls across 4 themes
-- ============================================================

-- Framework ID: a0000001-0000-0000-0000-000000000001
-- Domain IDs: d0000001-0001-0000-0000-000000000001 to 004

-- ============================================================
-- A.5 ORGANISATIONAL CONTROLS (37 controls)
-- ============================================================
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.1','Policies for information security','directive','administrative',1),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.2','Information security roles and responsibilities','directive','administrative',2),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.3','Segregation of duties','preventive','administrative',3),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.4','Management responsibilities','directive','administrative',4),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.5','Contact with authorities','preventive','administrative',5),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.6','Contact with special interest groups','preventive','administrative',6),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.7','Threat intelligence','preventive','administrative',7),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.8','Information security in project management','preventive','administrative',8),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.9','Inventory of information and other associated assets','preventive','administrative',9),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.10','Acceptable use of information and other associated assets','preventive','administrative',10),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.11','Return of assets','preventive','administrative',11),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.12','Classification of information','preventive','administrative',12),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.13','Labelling of information','preventive','administrative',13),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.14','Information transfer','preventive','administrative',14),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.15','Access control','preventive','administrative',15),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.16','Identity management','preventive','administrative',16),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.17','Authentication information','preventive','administrative',17),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.18','Access rights','preventive','administrative',18),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.19','Information security in supplier relationships','preventive','administrative',19),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.20','Addressing information security within supplier agreements','preventive','administrative',20),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.21','Managing information security in the ICT supply chain','preventive','administrative',21),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.22','Monitoring, review and change management of supplier services','detective','administrative',22),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.23','Information security for use of cloud services','preventive','administrative',23),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.24','Information security incident management planning and preparation','corrective','administrative',24),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.25','Assessment and decision on information security events','detective','administrative',25),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.26','Response to information security incidents','corrective','administrative',26),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.27','Learning from information security incidents','preventive','administrative',27),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.28','Collection of evidence','detective','administrative',28),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.29','Information security during disruption','preventive','administrative',29),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.30','ICT readiness for business continuity','preventive','administrative',30),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.31','Legal, statutory, regulatory and contractual requirements','preventive','administrative',31),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.32','Intellectual property rights','preventive','administrative',32),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.33','Protection of records','preventive','administrative',33),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.34','Privacy and protection of personal information','preventive','administrative',34),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.35','Independent review of information security','detective','administrative',35),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.36','Compliance with policies, rules and standards','detective','administrative',36),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000001','A.5.37','Documented operating procedures','preventive','administrative',37);

-- ============================================================
-- A.6 PEOPLE CONTROLS (8 controls)
-- ============================================================
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000002','A.6.1','Screening','preventive','administrative',38),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000002','A.6.2','Terms and conditions of employment','preventive','administrative',39),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000002','A.6.3','Information security awareness, education and training','preventive','administrative',40),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000002','A.6.4','Disciplinary process','corrective','administrative',41),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000002','A.6.5','Responsibilities after termination or change of employment','preventive','administrative',42),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000002','A.6.6','Confidentiality or non-disclosure agreements','preventive','administrative',43),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000002','A.6.7','Remote working','preventive','administrative',44),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000002','A.6.8','Information security event reporting','detective','administrative',45);

-- ============================================================
-- A.7 PHYSICAL CONTROLS (14 controls)
-- ============================================================
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000003','A.7.1','Physical security perimeters','preventive','physical',46),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000003','A.7.2','Physical entry','preventive','physical',47),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000003','A.7.3','Securing offices, rooms and facilities','preventive','physical',48),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000003','A.7.4','Physical security monitoring','detective','physical',49),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000003','A.7.5','Protecting against physical and environmental threats','preventive','physical',50),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000003','A.7.6','Working in secure areas','preventive','physical',51),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000003','A.7.7','Clear desk and clear screen','preventive','physical',52),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000003','A.7.8','Equipment siting and protection','preventive','physical',53),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000003','A.7.9','Security of assets off-premises','preventive','physical',54),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000003','A.7.10','Storage media','preventive','physical',55),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000003','A.7.11','Supporting utilities','preventive','physical',56),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000003','A.7.12','Cabling security','preventive','physical',57),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000003','A.7.13','Equipment maintenance','preventive','physical',58),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000003','A.7.14','Secure disposal or re-use of equipment','preventive','physical',59);

-- ============================================================
-- A.8 TECHNOLOGICAL CONTROLS (34 controls)
-- ============================================================
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000004','A.8.1','User endpoint devices','preventive','technical',60),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000004','A.8.2','Privileged access rights','preventive','technical',61),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000004','A.8.3','Information access restriction','preventive','technical',62),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000004','A.8.4','Access to source code','preventive','technical',63),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000004','A.8.5','Secure authentication','preventive','technical',64),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000004','A.8.6','Capacity management','preventive','technical',65),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000004','A.8.7','Protection against malware','preventive','technical',66),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000004','A.8.8','Management of technical vulnerabilities','preventive','technical',67),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000004','A.8.9','Configuration management','preventive','technical',68),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000004','A.8.10','Information deletion','preventive','technical',69),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000004','A.8.11','Data masking','preventive','technical',70),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000004','A.8.12','Data leakage prevention','preventive','technical',71),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000004','A.8.13','Information backup','preventive','technical',72),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000004','A.8.14','Redundancy of information processing facilities','preventive','technical',73),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000004','A.8.15','Logging','detective','technical',74),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000004','A.8.16','Monitoring activities','detective','technical',75),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000004','A.8.17','Clock synchronisation','preventive','technical',76),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000004','A.8.18','Use of privileged utility programs','preventive','technical',77),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000004','A.8.19','Installation of software on operational systems','preventive','technical',78),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000004','A.8.20','Networks security','preventive','technical',79),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000004','A.8.21','Security of network services','preventive','technical',80),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000004','A.8.22','Segregation of networks','preventive','technical',81),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000004','A.8.23','Web filtering','preventive','technical',82),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000004','A.8.24','Use of cryptography','preventive','technical',83),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000004','A.8.25','Secure development life cycle','preventive','technical',84),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000004','A.8.26','Application security requirements','preventive','technical',85),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000004','A.8.27','Secure system architecture and engineering principles','preventive','technical',86),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000004','A.8.28','Secure coding','preventive','technical',87),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000004','A.8.29','Security testing in development and acceptance','detective','technical',88),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000004','A.8.30','Outsourced development','preventive','technical',89),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000004','A.8.31','Separation of development, test and production environments','preventive','technical',90),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000004','A.8.32','Change management','preventive','technical',91),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000004','A.8.33','Test information','preventive','technical',92),
('a0000001-0000-0000-0000-000000000001','d0000001-0001-0000-0000-000000000004','A.8.34','Protection of information systems during audit testing','preventive','technical',93);
