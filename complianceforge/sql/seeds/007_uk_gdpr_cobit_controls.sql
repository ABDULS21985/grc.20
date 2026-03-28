-- ============================================================
-- ComplianceForge — Seed: UK GDPR Articles
-- ============================================================

-- CH2: Principles
INSERT INTO framework_controls (framework_id, domain_id, code, title, description, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000001','Art.5(1)(a)','Lawfulness, fairness and transparency','Personal data shall be processed lawfully, fairly and in a transparent manner.','directive','administrative',1),
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000001','Art.5(1)(b)','Purpose limitation','Collected for specified, explicit and legitimate purposes.','directive','administrative',2),
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000001','Art.5(1)(c)','Data minimisation','Adequate, relevant and limited to what is necessary.','directive','administrative',3),
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000001','Art.5(1)(d)','Accuracy','Accurate and where necessary kept up to date.','directive','administrative',4),
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000001','Art.5(1)(e)','Storage limitation','Kept no longer than is necessary for the purposes.','directive','administrative',5),
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000001','Art.5(1)(f)','Integrity and confidentiality','Processed with appropriate security measures.','directive','technical',6),
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000001','Art.5(2)','Accountability','The controller shall be responsible for and be able to demonstrate compliance.','directive','administrative',7);

-- CH3: Data Subject Rights
INSERT INTO framework_controls (framework_id, domain_id, code, title, description, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000002','Art.12','Transparent information and communication','Provide information in a concise, transparent, intelligible and easily accessible form.','directive','administrative',8),
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000002','Art.13','Information to be provided where data collected from the data subject','Privacy notice requirements when data collected directly.','directive','administrative',9),
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000002','Art.14','Information where data not obtained from the data subject','Privacy notice requirements when data obtained indirectly.','directive','administrative',10),
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000002','Art.15','Right of access by the data subject','Data subjects have the right to obtain confirmation of processing and access to data.','directive','administrative',11),
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000002','Art.16','Right to rectification','Right to have inaccurate personal data corrected without undue delay.','corrective','administrative',12),
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000002','Art.17','Right to erasure (right to be forgotten)','Right to have personal data erased under certain circumstances.','corrective','technical',13),
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000002','Art.18','Right to restriction of processing','Right to restrict processing under certain circumstances.','preventive','technical',14),
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000002','Art.19','Notification obligation regarding rectification or erasure','Communicate any rectification or erasure to each recipient.','corrective','administrative',15),
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000002','Art.20','Right to data portability','Right to receive data in a structured, commonly used and machine-readable format.','directive','technical',16),
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000002','Art.21','Right to object','Right to object to processing based on legitimate interests or direct marketing.','directive','administrative',17),
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000002','Art.22','Automated individual decision-making including profiling','Right not to be subject to solely automated decision-making with legal effects.','directive','technical',18);

-- CH4: Controller and Processor Obligations
INSERT INTO framework_controls (framework_id, domain_id, code, title, description, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000003','Art.24','Responsibility of the controller','Implement appropriate technical and organisational measures.','directive','administrative',19),
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000003','Art.25','Data protection by design and by default','Implement data protection principles at the design stage.','preventive','technical',20),
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000003','Art.26','Joint controllers','Determine respective responsibilities for compliance.','directive','administrative',21),
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000003','Art.28','Processor','Processing by a processor shall be governed by a contract.','directive','administrative',22),
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000003','Art.30','Records of processing activities','Maintain a record of processing activities (ROPA).','preventive','administrative',23),
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000003','Art.32','Security of processing','Implement appropriate technical and organisational security measures.','preventive','technical',24),
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000003','Art.33','Notification of breach to supervisory authority','Notify within 72 hours of becoming aware of a personal data breach.','corrective','administrative',25),
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000003','Art.34','Communication of breach to data subject','Communicate breach to data subjects when high risk to rights and freedoms.','corrective','administrative',26),
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000003','Art.35','Data protection impact assessment','Carry out DPIA where processing is likely to result in high risk.','preventive','administrative',27),
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000003','Art.36','Prior consultation','Consult the supervisory authority before processing where DPIA indicates high risk.','preventive','administrative',28),
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000003','Art.37','Designation of the data protection officer','Designate a DPO where required.','directive','administrative',29),
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000003','Art.38','Position of the data protection officer','Ensure the DPO is involved in all issues relating to data protection.','directive','administrative',30),
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000003','Art.39','Tasks of the data protection officer','DPO tasks include informing, advising, monitoring, and cooperating with authorities.','directive','administrative',31);

-- CH5: International Transfers
INSERT INTO framework_controls (framework_id, domain_id, code, title, description, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000004','Art.44','General principle for transfers','Transfer only if conditions in Chapter V are complied with.','directive','administrative',32),
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000004','Art.45','Transfers on basis of adequacy decision','Transfer to countries with adequate level of protection.','directive','administrative',33),
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000004','Art.46','Transfers subject to appropriate safeguards','Standard contractual clauses, binding corporate rules, etc.','preventive','administrative',34),
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000004','Art.47','Binding corporate rules','BCRs approved by competent supervisory authority.','preventive','administrative',35),
('a0000001-0000-0000-0000-000000000002','d0000002-0001-0000-0000-000000000004','Art.49','Derogations for specific situations','Transfer allowed in specific situations without safeguards.','directive','administrative',36);

-- ============================================================
-- COBIT 2019 — 40 Governance & Management Objectives
-- ============================================================

-- EDM: Evaluate, Direct and Monitor (5)
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000001','EDM01','Ensured Governance Framework Setting and Maintenance','directive','management',1),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000001','EDM02','Ensured Benefits Delivery','directive','management',2),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000001','EDM03','Ensured Risk Optimisation','directive','management',3),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000001','EDM04','Ensured Resource Optimisation','directive','management',4),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000001','EDM05','Ensured Stakeholder Engagement','directive','management',5);

-- APO: Align, Plan and Organize (14)
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000002','APO01','Managed I&T Management Framework','directive','administrative',6),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000002','APO02','Managed Strategy','directive','management',7),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000002','APO03','Managed Enterprise Architecture','directive','technical',8),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000002','APO04','Managed Innovation','directive','management',9),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000002','APO05','Managed Portfolio','directive','management',10),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000002','APO06','Managed Budget and Costs','preventive','administrative',11),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000002','APO07','Managed Human Resources','preventive','administrative',12),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000002','APO08','Managed Relationships','directive','administrative',13),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000002','APO09','Managed Service Agreements','directive','administrative',14),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000002','APO10','Managed Vendors','preventive','administrative',15),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000002','APO11','Managed Quality','detective','administrative',16),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000002','APO12','Managed Risk','preventive','administrative',17),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000002','APO13','Managed Security','preventive','technical',18),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000002','APO14','Managed Data','preventive','administrative',19);

-- BAI: Build, Acquire and Implement (11)
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000003','BAI01','Managed Programs','directive','management',20),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000003','BAI02','Managed Requirements Definition','directive','administrative',21),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000003','BAI03','Managed Solutions Identification and Build','preventive','technical',22),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000003','BAI04','Managed Availability and Capacity','preventive','technical',23),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000003','BAI05','Managed Organisational Change','directive','administrative',24),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000003','BAI06','Managed IT Changes','preventive','technical',25),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000003','BAI07','Managed IT Change Acceptance and Transitioning','detective','technical',26),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000003','BAI08','Managed Knowledge','preventive','administrative',27),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000003','BAI09','Managed Assets','preventive','administrative',28),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000003','BAI10','Managed Configuration','preventive','technical',29),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000003','BAI11','Managed Projects','directive','management',30);

-- DSS: Deliver, Service and Support (6)
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000004','DSS01','Managed Operations','preventive','technical',31),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000004','DSS02','Managed Service Requests and Incidents','corrective','administrative',32),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000004','DSS03','Managed Problems','corrective','administrative',33),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000004','DSS04','Managed Continuity','preventive','administrative',34),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000004','DSS05','Managed Security Services','preventive','technical',35),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000004','DSS06','Managed Business Process Controls','preventive','administrative',36);

-- MEA: Monitor, Evaluate and Assess (4)
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000005','MEA01','Managed Performance and Conformance Monitoring','detective','administrative',37),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000005','MEA02','Managed System of Internal Control','detective','administrative',38),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000005','MEA03','Managed Compliance with External Requirements','detective','administrative',39),
('a0000001-0000-0000-0000-000000000009','d0000009-0001-0000-0000-000000000005','MEA04','Managed Assurance','detective','administrative',40);
