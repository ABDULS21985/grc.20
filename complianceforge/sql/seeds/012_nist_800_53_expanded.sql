-- ============================================================
-- ComplianceForge — Seed: NIST SP 800-53 Rev 5 — Expanded Controls
-- Additional families: PL, PT, SA + depth on existing families
-- ============================================================

-- PL: Planning
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000012','PL-1','Policy and Procedures','directive','administrative',130),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000012','PL-2','System Security and Privacy Plans','directive','administrative',131),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000012','PL-4','Rules of Behaviour','directive','administrative',132),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000012','PL-7','Concept of Operations','directive','management',133),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000012','PL-8','Security and Privacy Architectures','directive','technical',134),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000012','PL-10','Baseline Selection','directive','administrative',135),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000012','PL-11','Baseline Tailoring','directive','administrative',136);

-- PT: Personally Identifiable Information Processing
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000015','PT-1','Policy and Procedures','directive','administrative',137),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000015','PT-2','Authority to Process Personally Identifiable Information','directive','administrative',138),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000015','PT-3','Personally Identifiable Information Processing Purposes','directive','administrative',139),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000015','PT-4','Consent','preventive','administrative',140),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000015','PT-5','Privacy Notice','directive','administrative',141),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000015','PT-6','System of Records Notice','directive','administrative',142),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000015','PT-7','Specific Categories of Personally Identifiable Information','preventive','administrative',143),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000015','PT-8','Computer Matching Requirements','preventive','administrative',144);

-- SA: System and Services Acquisition
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000017','SA-1','Policy and Procedures','directive','administrative',145),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000017','SA-2','Allocation of Resources','directive','management',146),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000017','SA-3','System Development Life Cycle','preventive','technical',147),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000017','SA-4','Acquisition Process','preventive','administrative',148),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000017','SA-5','System Documentation','preventive','administrative',149),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000017','SA-8','Security and Privacy Engineering Principles','preventive','technical',150),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000017','SA-9','External System Services','preventive','administrative',151),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000017','SA-10','Developer Configuration Management','preventive','technical',152),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000017','SA-11','Developer Testing and Evaluation','detective','technical',153),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000017','SA-15','Development Process Standards and Tools','preventive','technical',154),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000017','SA-17','Developer Security and Privacy Architecture and Design','preventive','technical',155),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000017','SA-22','Unsupported System Components','preventive','administrative',156);

-- PM: Program Management
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000013','PM-1','Information Security Program Plan','directive','management',157),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000013','PM-2','Information Security Program Leadership Role','directive','management',158),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000013','PM-3','Information Security and Privacy Resources','directive','management',159),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000013','PM-4','Plan of Action and Milestones Process','corrective','administrative',160),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000013','PM-5','System Inventory','preventive','administrative',161),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000013','PM-6','Measures of Performance','detective','management',162),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000013','PM-7','Enterprise Architecture','directive','management',163),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000013','PM-8','Critical Infrastructure Plan','preventive','management',164),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000013','PM-9','Risk Management Strategy','directive','management',165),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000013','PM-10','Authorisation Process','directive','administrative',166),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000013','PM-11','Mission and Business Process Definition','directive','management',167),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000013','PM-13','Security and Privacy Workforce','preventive','administrative',168),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000013','PM-14','Testing Training and Monitoring','detective','administrative',169),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000013','PM-15','Security and Privacy Groups and Associations','preventive','administrative',170),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000013','PM-16','Threat Awareness Program','preventive','administrative',171),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000013','PM-17','Protecting Controlled Unclassified Information on External Systems','preventive','technical',172),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000013','PM-20','Dissemination of Privacy Program Information','directive','administrative',173),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000013','PM-21','Accounting of Disclosures','detective','administrative',174),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000013','PM-22','Personally Identifiable Information Quality Management','preventive','administrative',175),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000013','PM-25','Minimisation of Personally Identifiable Information Used in Testing','preventive','technical',176),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000013','PM-26','Complaint Management','corrective','administrative',177),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000013','PM-27','Privacy Reporting','detective','administrative',178),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000013','PM-28','Risk Framing','directive','management',179),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000013','PM-30','Supply Chain Risk Management Strategy','directive','management',180),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000013','PM-31','Continuous Monitoring Strategy','detective','management',181),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000013','PM-32','Purposing','directive','management',182);

-- Additional SC (System & Communications Protection) enhancements
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000018','SC-2','Separation of System and User Functionality','preventive','technical',183),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000018','SC-3','Security Function Isolation','preventive','technical',184),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000018','SC-4','Information in Shared System Resources','preventive','technical',185),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000018','SC-10','Network Disconnect','preventive','technical',186),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000018','SC-15','Collaborative Computing Devices and Applications','preventive','technical',187),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000018','SC-17','Public Key Infrastructure Certificates','preventive','technical',188),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000018','SC-18','Mobile Code','preventive','technical',189),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000018','SC-20','Secure Name Address Resolution Service (Authoritative Source)','preventive','technical',190),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000018','SC-21','Secure Name Address Resolution Service (Recursive or Caching Resolver)','preventive','technical',191),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000018','SC-22','Architecture and Provisioning for Name Address Resolution Service','preventive','technical',192),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000018','SC-39','Process Isolation','preventive','technical',193);

-- Additional SI (System & Information Integrity) enhancements
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000019','SI-6','Security and Privacy Function Verification','detective','technical',194),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000019','SI-8','Spam Protection','preventive','technical',195),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000019','SI-11','Error Handling','preventive','technical',196),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000019','SI-16','Memory Protection','preventive','technical',197);
