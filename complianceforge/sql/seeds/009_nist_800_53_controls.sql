-- ============================================================
-- ComplianceForge — Seed: NIST SP 800-53 Rev 5 Controls
-- Key controls from all 20 families (representative subset)
-- ============================================================

-- AC: Access Control
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000001','AC-1','Policy and Procedures','directive','administrative',1),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000001','AC-2','Account Management','preventive','technical',2),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000001','AC-3','Access Enforcement','preventive','technical',3),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000001','AC-4','Information Flow Enforcement','preventive','technical',4),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000001','AC-5','Separation of Duties','preventive','administrative',5),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000001','AC-6','Least Privilege','preventive','technical',6),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000001','AC-7','Unsuccessful Logon Attempts','preventive','technical',7),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000001','AC-8','System Use Notification','preventive','technical',8),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000001','AC-11','Device Lock','preventive','technical',9),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000001','AC-14','Permitted Actions Without Identification or Authentication','preventive','technical',10),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000001','AC-17','Remote Access','preventive','technical',11),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000001','AC-18','Wireless Access','preventive','technical',12),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000001','AC-19','Access Control for Mobile Devices','preventive','technical',13),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000001','AC-20','Use of External Systems','preventive','administrative',14),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000001','AC-22','Publicly Accessible Content','preventive','administrative',15);

-- AT: Awareness and Training
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000002','AT-1','Policy and Procedures','directive','administrative',16),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000002','AT-2','Literacy Training and Awareness','preventive','administrative',17),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000002','AT-3','Role-Based Training','preventive','administrative',18),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000002','AT-4','Training Records','detective','administrative',19);

-- AU: Audit and Accountability
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000003','AU-1','Policy and Procedures','directive','administrative',20),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000003','AU-2','Event Logging','detective','technical',21),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000003','AU-3','Content of Audit Records','detective','technical',22),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000003','AU-4','Audit Log Storage Capacity','preventive','technical',23),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000003','AU-5','Response to Audit Logging Process Failures','corrective','technical',24),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000003','AU-6','Audit Record Review Analysis and Reporting','detective','administrative',25),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000003','AU-8','Time Stamps','preventive','technical',26),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000003','AU-9','Protection of Audit Information','preventive','technical',27),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000003','AU-11','Audit Record Retention','preventive','administrative',28),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000003','AU-12','Audit Record Generation','detective','technical',29);

-- CA: Assessment Authorization and Monitoring
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000004','CA-1','Policy and Procedures','directive','administrative',30),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000004','CA-2','Control Assessments','detective','administrative',31),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000004','CA-3','Information Exchange','preventive','administrative',32),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000004','CA-5','Plan of Action and Milestones','corrective','administrative',33),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000004','CA-6','Authorisation','directive','administrative',34),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000004','CA-7','Continuous Monitoring','detective','technical',35),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000004','CA-8','Penetration Testing','detective','technical',36),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000004','CA-9','Internal System Connections','preventive','technical',37);

-- CM: Configuration Management
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000005','CM-1','Policy and Procedures','directive','administrative',38),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000005','CM-2','Baseline Configuration','preventive','technical',39),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000005','CM-3','Configuration Change Control','preventive','technical',40),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000005','CM-4','Impact Analyses','detective','administrative',41),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000005','CM-5','Access Restrictions for Change','preventive','technical',42),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000005','CM-6','Configuration Settings','preventive','technical',43),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000005','CM-7','Least Functionality','preventive','technical',44),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000005','CM-8','System Component Inventory','preventive','administrative',45),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000005','CM-10','Software Usage Restrictions','preventive','administrative',46),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000005','CM-11','User-Installed Software','preventive','technical',47);

-- CP: Contingency Planning
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000006','CP-1','Policy and Procedures','directive','administrative',48),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000006','CP-2','Contingency Plan','preventive','administrative',49),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000006','CP-3','Contingency Training','preventive','administrative',50),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000006','CP-4','Contingency Plan Testing','detective','administrative',51),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000006','CP-6','Alternate Storage Site','preventive','physical',52),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000006','CP-7','Alternate Processing Site','preventive','physical',53),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000006','CP-9','System Backup','preventive','technical',54),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000006','CP-10','System Recovery and Reconstitution','corrective','technical',55);

-- IA: Identification and Authentication
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000007','IA-1','Policy and Procedures','directive','administrative',56),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000007','IA-2','Identification and Authentication (Organisational Users)','preventive','technical',57),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000007','IA-3','Device Identification and Authentication','preventive','technical',58),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000007','IA-4','Identifier Management','preventive','technical',59),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000007','IA-5','Authenticator Management','preventive','technical',60),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000007','IA-6','Authentication Feedback','preventive','technical',61),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000007','IA-8','Identification and Authentication (Non-Organisational Users)','preventive','technical',62),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000007','IA-11','Re-Authentication','preventive','technical',63),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000007','IA-12','Identity Proofing','preventive','administrative',64);

-- IR: Incident Response
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000008','IR-1','Policy and Procedures','directive','administrative',65),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000008','IR-2','Incident Response Training','preventive','administrative',66),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000008','IR-3','Incident Response Testing','detective','administrative',67),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000008','IR-4','Incident Handling','corrective','administrative',68),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000008','IR-5','Incident Monitoring','detective','technical',69),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000008','IR-6','Incident Reporting','corrective','administrative',70),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000008','IR-7','Incident Response Assistance','corrective','administrative',71),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000008','IR-8','Incident Response Plan','preventive','administrative',72);

-- PE: Physical and Environmental Protection
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000011','PE-1','Policy and Procedures','directive','administrative',73),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000011','PE-2','Physical Access Authorisations','preventive','physical',74),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000011','PE-3','Physical Access Control','preventive','physical',75),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000011','PE-6','Monitoring Physical Access','detective','physical',76),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000011','PE-8','Visitor Access Records','detective','physical',77),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000011','PE-13','Fire Protection','preventive','physical',78),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000011','PE-14','Environmental Controls','preventive','physical',79),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000011','PE-15','Water Damage Protection','preventive','physical',80);

-- RA: Risk Assessment
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000016','RA-1','Policy and Procedures','directive','administrative',81),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000016','RA-2','Security Categorisation','preventive','administrative',82),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000016','RA-3','Risk Assessment','detective','administrative',83),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000016','RA-5','Vulnerability Monitoring and Scanning','detective','technical',84),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000016','RA-7','Risk Response','corrective','administrative',85);

-- SC: System and Communications Protection
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000018','SC-1','Policy and Procedures','directive','administrative',86),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000018','SC-5','Denial-of-Service Protection','preventive','technical',87),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000018','SC-7','Boundary Protection','preventive','technical',88),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000018','SC-8','Transmission Confidentiality and Integrity','preventive','technical',89),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000018','SC-12','Cryptographic Key Establishment and Management','preventive','technical',90),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000018','SC-13','Cryptographic Protection','preventive','technical',91),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000018','SC-23','Session Authenticity','preventive','technical',92),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000018','SC-28','Protection of Information at Rest','preventive','technical',93);

-- SI: System and Information Integrity
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000019','SI-1','Policy and Procedures','directive','administrative',94),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000019','SI-2','Flaw Remediation','corrective','technical',95),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000019','SI-3','Malicious Code Protection','preventive','technical',96),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000019','SI-4','System Monitoring','detective','technical',97),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000019','SI-5','Security Alerts Advisories and Directives','detective','administrative',98),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000019','SI-7','Software Firmware and Information Integrity','detective','technical',99),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000019','SI-10','Information Input Validation','preventive','technical',100),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000019','SI-12','Information Management and Retention','preventive','administrative',101);

-- SR: Supply Chain Risk Management
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000020','SR-1','Policy and Procedures','directive','administrative',102),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000020','SR-2','Supply Chain Risk Management Plan','preventive','administrative',103),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000020','SR-3','Supply Chain Controls and Processes','preventive','administrative',104),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000020','SR-5','Acquisition Strategies Tools and Methods','preventive','administrative',105),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000020','SR-6','Supplier Assessments and Reviews','detective','administrative',106),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000020','SR-11','Component Authenticity','detective','technical',107);

-- PS: Personnel Security
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000014','PS-1','Policy and Procedures','directive','administrative',108),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000014','PS-2','Position Risk Designation','preventive','administrative',109),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000014','PS-3','Personnel Screening','preventive','administrative',110),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000014','PS-4','Personnel Termination','preventive','administrative',111),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000014','PS-5','Personnel Transfer','preventive','administrative',112),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000014','PS-6','Access Agreements','preventive','administrative',113),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000014','PS-7','External Personnel Security','preventive','administrative',114),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000014','PS-8','Personnel Sanctions','corrective','administrative',115);

-- MA: Maintenance
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000009','MA-1','Policy and Procedures','directive','administrative',116),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000009','MA-2','Controlled Maintenance','preventive','physical',117),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000009','MA-4','Non-Local Maintenance','preventive','technical',118),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000009','MA-5','Maintenance Personnel','preventive','administrative',119);

-- MP: Media Protection
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000010','MP-1','Policy and Procedures','directive','administrative',120),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000010','MP-2','Media Access','preventive','physical',121),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000010','MP-4','Media Storage','preventive','physical',122),
('a0000001-0000-0000-0000-000000000005','d0000005-0001-0000-0000-000000000010','MP-6','Media Sanitisation','preventive','physical',123);
