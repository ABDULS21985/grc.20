-- ============================================================
-- ComplianceForge — Seed: NCSC CAF Principles
-- 4 Objectives, 14 Principles with contributing outcomes
-- ============================================================

-- A: Managing Security Risk
INSERT INTO framework_controls (framework_id, domain_id, code, title, description, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000003','d0000003-0001-0000-0000-000000000001','A1','Governance','The organisation has appropriate management policies and processes in place to govern its approach to the security of network and information systems.','directive','management',1),
('a0000001-0000-0000-0000-000000000003','d0000003-0001-0000-0000-000000000001','A2','Risk Management','The organisation takes appropriate steps to identify, assess and understand security risks to the network and information systems that support the delivery of essential services.','preventive','administrative',2),
('a0000001-0000-0000-0000-000000000003','d0000003-0001-0000-0000-000000000001','A3','Asset Management','Everything required to deliver, maintain or support networks and information systems necessary for the provision of essential services is determined and understood.','preventive','administrative',3),
('a0000001-0000-0000-0000-000000000003','d0000003-0001-0000-0000-000000000001','A4','Supply Chain','The organisation understands and manages security risks to networks and information systems that arise as a result of dependencies on external suppliers.','preventive','administrative',4);

-- B: Protecting Against Cyber Attack
INSERT INTO framework_controls (framework_id, domain_id, code, title, description, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000003','d0000003-0001-0000-0000-000000000002','B1','Service Protection Policies and Processes','The organisation defines, implements, communicates and enforces appropriate policies and processes to direct its overall approach to securing systems and data.','directive','administrative',5),
('a0000001-0000-0000-0000-000000000003','d0000003-0001-0000-0000-000000000002','B2','Identity and Access Control','The organisation understands, documents and manages access to network and information systems that support the delivery of essential services.','preventive','technical',6),
('a0000001-0000-0000-0000-000000000003','d0000003-0001-0000-0000-000000000002','B3','Data Security','Data stored or transmitted electronically is protected from actions that may cause loss of availability, confidentiality or integrity.','preventive','technical',7),
('a0000001-0000-0000-0000-000000000003','d0000003-0001-0000-0000-000000000002','B4','System Security','Network and information systems that support the delivery of essential services are protected from cyber attack.','preventive','technical',8),
('a0000001-0000-0000-0000-000000000003','d0000003-0001-0000-0000-000000000002','B5','Resilient Networks and Systems','The organisation builds resilience against cyber attack.','preventive','technical',9),
('a0000001-0000-0000-0000-000000000003','d0000003-0001-0000-0000-000000000002','B6','Staff Awareness and Training','Staff have appropriate awareness, knowledge and skills to carry out their organisational roles effectively in relation to the security of network and information systems.','preventive','administrative',10);

-- C: Detecting Cyber Security Events
INSERT INTO framework_controls (framework_id, domain_id, code, title, description, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000003','d0000003-0001-0000-0000-000000000003','C1','Security Monitoring','The organisation monitors the security status of the networks and systems that support the delivery of essential services in order to detect potential security problems.','detective','technical',11),
('a0000001-0000-0000-0000-000000000003','d0000003-0001-0000-0000-000000000003','C2','Proactive Security Event Discovery','The organisation detects, within networks and information systems, malicious activity affecting or with the potential to affect the delivery of essential services.','detective','technical',12);

-- D: Minimising the Impact of Cyber Security Incidents
INSERT INTO framework_controls (framework_id, domain_id, code, title, description, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000003','d0000003-0001-0000-0000-000000000004','D1','Response and Recovery Planning','There are well-defined and tested incident management processes in place that aim to ensure continuity of essential services in the event of system or service failure.','corrective','administrative',13),
('a0000001-0000-0000-0000-000000000003','d0000003-0001-0000-0000-000000000004','D2','Lessons Learned','When a cyber security incident occurs the organisation learns lessons including where appropriate updating security measures.','corrective','administrative',14);

-- ============================================================
-- ITIL 4 — 34 Management Practices
-- ============================================================

-- General Management Practices (14)
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000008','d0000008-0001-0000-0000-000000000001','ITIL-GP01','Architecture Management','directive','management',1),
('a0000001-0000-0000-0000-000000000008','d0000008-0001-0000-0000-000000000001','ITIL-GP02','Continual Improvement','corrective','administrative',2),
('a0000001-0000-0000-0000-000000000008','d0000008-0001-0000-0000-000000000001','ITIL-GP03','Information Security Management','preventive','technical',3),
('a0000001-0000-0000-0000-000000000008','d0000008-0001-0000-0000-000000000001','ITIL-GP04','Knowledge Management','preventive','administrative',4),
('a0000001-0000-0000-0000-000000000008','d0000008-0001-0000-0000-000000000001','ITIL-GP05','Measurement and Reporting','detective','administrative',5),
('a0000001-0000-0000-0000-000000000008','d0000008-0001-0000-0000-000000000001','ITIL-GP06','Organisational Change Management','directive','administrative',6),
('a0000001-0000-0000-0000-000000000008','d0000008-0001-0000-0000-000000000001','ITIL-GP07','Portfolio Management','directive','management',7),
('a0000001-0000-0000-0000-000000000008','d0000008-0001-0000-0000-000000000001','ITIL-GP08','Project Management','directive','management',8),
('a0000001-0000-0000-0000-000000000008','d0000008-0001-0000-0000-000000000001','ITIL-GP09','Relationship Management','directive','administrative',9),
('a0000001-0000-0000-0000-000000000008','d0000008-0001-0000-0000-000000000001','ITIL-GP10','Risk Management','preventive','administrative',10),
('a0000001-0000-0000-0000-000000000008','d0000008-0001-0000-0000-000000000001','ITIL-GP11','Service Financial Management','preventive','administrative',11),
('a0000001-0000-0000-0000-000000000008','d0000008-0001-0000-0000-000000000001','ITIL-GP12','Strategy Management','directive','management',12),
('a0000001-0000-0000-0000-000000000008','d0000008-0001-0000-0000-000000000001','ITIL-GP13','Supplier Management','preventive','administrative',13),
('a0000001-0000-0000-0000-000000000008','d0000008-0001-0000-0000-000000000001','ITIL-GP14','Workforce and Talent Management','preventive','administrative',14);

-- Service Management Practices (17)
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000008','d0000008-0001-0000-0000-000000000002','ITIL-SM01','Availability Management','preventive','technical',15),
('a0000001-0000-0000-0000-000000000008','d0000008-0001-0000-0000-000000000002','ITIL-SM02','Business Analysis','directive','administrative',16),
('a0000001-0000-0000-0000-000000000008','d0000008-0001-0000-0000-000000000002','ITIL-SM03','Capacity and Performance Management','preventive','technical',17),
('a0000001-0000-0000-0000-000000000008','d0000008-0001-0000-0000-000000000002','ITIL-SM04','Change Enablement','preventive','administrative',18),
('a0000001-0000-0000-0000-000000000008','d0000008-0001-0000-0000-000000000002','ITIL-SM05','Incident Management','corrective','administrative',19),
('a0000001-0000-0000-0000-000000000008','d0000008-0001-0000-0000-000000000002','ITIL-SM06','IT Asset Management','preventive','administrative',20),
('a0000001-0000-0000-0000-000000000008','d0000008-0001-0000-0000-000000000002','ITIL-SM07','Monitoring and Event Management','detective','technical',21),
('a0000001-0000-0000-0000-000000000008','d0000008-0001-0000-0000-000000000002','ITIL-SM08','Problem Management','corrective','administrative',22),
('a0000001-0000-0000-0000-000000000008','d0000008-0001-0000-0000-000000000002','ITIL-SM09','Release Management','preventive','technical',23),
('a0000001-0000-0000-0000-000000000008','d0000008-0001-0000-0000-000000000002','ITIL-SM10','Service Catalogue Management','directive','administrative',24),
('a0000001-0000-0000-0000-000000000008','d0000008-0001-0000-0000-000000000002','ITIL-SM11','Service Configuration Management','preventive','technical',25),
('a0000001-0000-0000-0000-000000000008','d0000008-0001-0000-0000-000000000002','ITIL-SM12','Service Continuity Management','preventive','administrative',26),
('a0000001-0000-0000-0000-000000000008','d0000008-0001-0000-0000-000000000002','ITIL-SM13','Service Design','directive','technical',27),
('a0000001-0000-0000-0000-000000000008','d0000008-0001-0000-0000-000000000002','ITIL-SM14','Service Desk','corrective','administrative',28),
('a0000001-0000-0000-0000-000000000008','d0000008-0001-0000-0000-000000000002','ITIL-SM15','Service Level Management','directive','administrative',29),
('a0000001-0000-0000-0000-000000000008','d0000008-0001-0000-0000-000000000002','ITIL-SM16','Service Request Management','corrective','administrative',30),
('a0000001-0000-0000-0000-000000000008','d0000008-0001-0000-0000-000000000002','ITIL-SM17','Service Validation and Testing','detective','technical',31);

-- Technical Management Practices (3)
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000008','d0000008-0001-0000-0000-000000000003','ITIL-TM01','Deployment Management','preventive','technical',32),
('a0000001-0000-0000-0000-000000000008','d0000008-0001-0000-0000-000000000003','ITIL-TM02','Infrastructure and Platform Management','preventive','technical',33),
('a0000001-0000-0000-0000-000000000008','d0000008-0001-0000-0000-000000000003','ITIL-TM03','Software Development and Management','preventive','technical',34);
