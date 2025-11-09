package seed

import (
	"fmt"
	"log"
)

// SeedEngineers creates 50 engineers across the organization
func SeedEngineers(data *SeedData) error {
	fmt.Println("\n4. Creating engineers (50 total)...")

	// Define all 50 engineers with realistic distribution
	engineers := []EngineerDef{
		// VP Engineering (1)
		{Name: "Sarah Chen", Email: "sarah.chen@example.com", Manager: "", Role: "Staff Engineer", TeamID: "Engineering",
			Identifiers: map[string]string{"github": "schen", "jira": "sarah.chen@example.com"}},

		// Backend Engineering - Director (1)
		{Name: "Marcus Johnson", Email: "marcus.johnson@example.com", Manager: "Sarah Chen", Role: "Staff Engineer", TeamID: "Backend Engineering",
			Identifiers: map[string]string{"github": "mjohnson", "jira": "marcus.johnson@example.com"}},

		// API Team - Manager + Engineers (10 total)
		{Name: "Elena Rodriguez", Email: "elena.rodriguez@example.com", Manager: "Marcus Johnson", Role: "Senior Engineer", TeamID: "API Team",
			Identifiers: map[string]string{"github": "erodriguez", "jira": "elena.rodriguez@example.com"}},
		{Name: "David Kim", Email: "david.kim@example.com", Manager: "Elena Rodriguez", Role: "Senior Engineer", TeamID: "API Team",
			Identifiers: map[string]string{"github": "dkim", "jira": "david.kim@example.com"}},
		{Name: "Aisha Patel", Email: "aisha.patel@example.com", Manager: "Elena Rodriguez", Role: "Senior Engineer", TeamID: "API Team",
			Identifiers: map[string]string{"github": "apatel", "jira": "aisha.patel@example.com"}},
		{Name: "James Wilson", Email: "james.wilson@example.com", Manager: "Elena Rodriguez", Role: "Senior Engineer", TeamID: "API Team",
			Identifiers: map[string]string{"github": "jwilson", "jira": "james.wilson@example.com"}},
		{Name: "Yuki Tanaka", Email: "yuki.tanaka@example.com", Manager: "Elena Rodriguez", Role: "Mid-Level Engineer", TeamID: "API Team",
			Identifiers: map[string]string{"github": "ytanaka", "jira": "yuki.tanaka@example.com"}},
		{Name: "Carlos Garcia", Email: "carlos.garcia@example.com", Manager: "Elena Rodriguez", Role: "Mid-Level Engineer", TeamID: "API Team",
			Identifiers: map[string]string{"github": "cgarcia", "jira": "carlos.garcia@example.com"}},
		{Name: "Zara Ahmed", Email: "zara.ahmed@example.com", Manager: "Elena Rodriguez", Role: "Senior Engineer", TeamID: "API Team",
			Identifiers: map[string]string{"github": "zahmed", "jira": "zara.ahmed@example.com"}},
		{Name: "Lucas Santos", Email: "lucas.santos@example.com", Manager: "Elena Rodriguez", Role: "Senior Engineer", TeamID: "API Team",
			Identifiers: map[string]string{"github": "lsantos", "jira": "lucas.santos@example.com"}},
		{Name: "Emma Thompson", Email: "emma.thompson@example.com", Manager: "Elena Rodriguez", Role: "Mid-Level Engineer", TeamID: "API Team",
			Identifiers: map[string]string{"github": "ethompson", "jira": "emma.thompson@example.com"}},
		{Name: "Omar Hassan", Email: "omar.hassan@example.com", Manager: "Elena Rodriguez", Role: "Mid-Level Engineer", TeamID: "API Team",
			Identifiers: map[string]string{"github": "ohassan", "jira": "omar.hassan@example.com"}},

		// Infrastructure Team - Manager + Engineers (10 total)
		{Name: "Priya Sharma", Email: "priya.sharma@example.com", Manager: "Marcus Johnson", Role: "Senior Engineer", TeamID: "Infrastructure",
			Identifiers: map[string]string{"github": "psharma", "jira": "priya.sharma@example.com"}},
		{Name: "Alex Thompson", Email: "alex.thompson@example.com", Manager: "Priya Sharma", Role: "Senior Engineer", TeamID: "Infrastructure",
			Identifiers: map[string]string{"github": "athompson", "jira": "alex.thompson@example.com"}},
		{Name: "Nina Okafor", Email: "nina.okafor@example.com", Manager: "Priya Sharma", Role: "Senior Engineer", TeamID: "Infrastructure",
			Identifiers: map[string]string{"github": "nokafor", "jira": "nina.okafor@example.com"}},
		{Name: "Ryan O'Connor", Email: "ryan.oconnor@example.com", Manager: "Priya Sharma", Role: "Senior Engineer", TeamID: "Infrastructure",
			Identifiers: map[string]string{"github": "roconnor", "jira": "ryan.oconnor@example.com"}},
		{Name: "Mei Lin", Email: "mei.lin@example.com", Manager: "Priya Sharma", Role: "Mid-Level Engineer", TeamID: "Infrastructure",
			Identifiers: map[string]string{"github": "mlin", "jira": "mei.lin@example.com"}},
		{Name: "Diego Martinez", Email: "diego.martinez@example.com", Manager: "Priya Sharma", Role: "Mid-Level Engineer", TeamID: "Infrastructure",
			Identifiers: map[string]string{"github": "dmartinez", "jira": "diego.martinez@example.com"}},
		{Name: "Fatima Ali", Email: "fatima.ali@example.com", Manager: "Priya Sharma", Role: "Senior Engineer", TeamID: "Infrastructure",
			Identifiers: map[string]string{"github": "fali", "jira": "fatima.ali@example.com"}},
		{Name: "Lars Johansson", Email: "lars.johansson@example.com", Manager: "Priya Sharma", Role: "Senior Engineer", TeamID: "Infrastructure",
			Identifiers: map[string]string{"github": "ljohansson", "jira": "lars.johansson@example.com"}},
		{Name: "Sophia Lee", Email: "sophia.lee@example.com", Manager: "Priya Sharma", Role: "Mid-Level Engineer", TeamID: "Infrastructure",
			Identifiers: map[string]string{"github": "slee", "jira": "sophia.lee@example.com"}},
		{Name: "Ahmed Khalil", Email: "ahmed.khalil@example.com", Manager: "Priya Sharma", Role: "Mid-Level Engineer", TeamID: "Infrastructure",
			Identifiers: map[string]string{"github": "akhalil", "jira": "ahmed.khalil@example.com"}},

		// Frontend Engineering - Director (1)
		{Name: "Maya Patel", Email: "maya.patel@example.com", Manager: "Sarah Chen", Role: "Staff Engineer", TeamID: "Frontend Engineering",
			Identifiers: map[string]string{"github": "mpatel", "jira": "maya.patel@example.com"}},

		// Web Team - Manager + Engineers (12 total)
		{Name: "Li Wei", Email: "li.wei@example.com", Manager: "Maya Patel", Role: "Senior Engineer", TeamID: "Web Team",
			Identifiers: map[string]string{"github": "lwei", "jira": "li.wei@example.com"}},
		{Name: "Isabella Rossi", Email: "isabella.rossi@example.com", Manager: "Li Wei", Role: "Senior Engineer", TeamID: "Web Team",
			Identifiers: map[string]string{"github": "irossi", "jira": "isabella.rossi@example.com"}},
		{Name: "Noah Brown", Email: "noah.brown@example.com", Manager: "Li Wei", Role: "Senior Engineer", TeamID: "Web Team",
			Identifiers: map[string]string{"github": "nbrown", "jira": "noah.brown@example.com"}},
		{Name: "Amara Osei", Email: "amara.osei@example.com", Manager: "Li Wei", Role: "Senior Engineer", TeamID: "Web Team",
			Identifiers: map[string]string{"github": "aosei", "jira": "amara.osei@example.com"}},
		{Name: "Kai Zhang", Email: "kai.zhang@example.com", Manager: "Li Wei", Role: "Mid-Level Engineer", TeamID: "Web Team",
			Identifiers: map[string]string{"github": "kzhang", "jira": "kai.zhang@example.com"}},
		{Name: "Olivia Smith", Email: "olivia.smith@example.com", Manager: "Li Wei", Role: "Mid-Level Engineer", TeamID: "Web Team",
			Identifiers: map[string]string{"github": "osmith", "jira": "olivia.smith@example.com"}},
		{Name: "Hassan Noor", Email: "hassan.noor@example.com", Manager: "Li Wei", Role: "Mid-Level Engineer", TeamID: "Web Team",
			Identifiers: map[string]string{"github": "hnoor", "jira": "hassan.noor@example.com"}},
		{Name: "Lucia Fernandez", Email: "lucia.fernandez@example.com", Manager: "Li Wei", Role: "Junior Engineer", TeamID: "Web Team",
			Identifiers: map[string]string{"github": "lfernandez", "jira": "lucia.fernandez@example.com"}},
		{Name: "Ethan Davis", Email: "ethan.davis@example.com", Manager: "Li Wei", Role: "Junior Engineer", TeamID: "Web Team",
			Identifiers: map[string]string{"github": "edavis", "jira": "ethan.davis@example.com"}},
		{Name: "Anya Ivanova", Email: "anya.ivanova@example.com", Manager: "Li Wei", Role: "Mid-Level Engineer", TeamID: "Web Team",
			Identifiers: map[string]string{"github": "aivanova", "jira": "anya.ivanova@example.com"}},
		{Name: "Jamal Washington", Email: "jamal.washington@example.com", Manager: "Li Wei", Role: "Mid-Level Engineer", TeamID: "Web Team",
			Identifiers: map[string]string{"github": "jwashington", "jira": "jamal.washington@example.com"}},
		{Name: "Chloe Dubois", Email: "chloe.dubois@example.com", Manager: "Li Wei", Role: "Senior Engineer", TeamID: "Web Team",
			Identifiers: map[string]string{"github": "cdubois", "jira": "chloe.dubois@example.com"}},

		// Mobile Team - Manager + Engineers (8 total)
		{Name: "Hiroshi Yamamoto", Email: "hiroshi.yamamoto@example.com", Manager: "Maya Patel", Role: "Senior Engineer", TeamID: "Mobile Team",
			Identifiers: map[string]string{"github": "hyamamoto", "jira": "hiroshi.yamamoto@example.com"}},
		{Name: "Sofia Lopez", Email: "sofia.lopez@example.com", Manager: "Hiroshi Yamamoto", Role: "Senior Engineer", TeamID: "Mobile Team",
			Identifiers: map[string]string{"github": "slopez", "jira": "sofia.lopez@example.com"}},
		{Name: "Daniel Cohen", Email: "daniel.cohen@example.com", Manager: "Hiroshi Yamamoto", Role: "Senior Engineer", TeamID: "Mobile Team",
			Identifiers: map[string]string{"github": "dcohen", "jira": "daniel.cohen@example.com"}},
		{Name: "Amina Diallo", Email: "amina.diallo@example.com", Manager: "Hiroshi Yamamoto", Role: "Mid-Level Engineer", TeamID: "Mobile Team",
			Identifiers: map[string]string{"github": "adiallo", "jira": "amina.diallo@example.com"}},
		{Name: "Mateo Silva", Email: "mateo.silva@example.com", Manager: "Hiroshi Yamamoto", Role: "Mid-Level Engineer", TeamID: "Mobile Team",
			Identifiers: map[string]string{"github": "msilva", "jira": "mateo.silva@example.com"}},
		{Name: "Leila Haddad", Email: "leila.haddad@example.com", Manager: "Hiroshi Yamamoto", Role: "Senior Engineer", TeamID: "Mobile Team",
			Identifiers: map[string]string{"github": "lhaddad", "jira": "leila.haddad@example.com"}},
		{Name: "Ivan Petrov", Email: "ivan.petrov@example.com", Manager: "Hiroshi Yamamoto", Role: "Senior Engineer", TeamID: "Mobile Team",
			Identifiers: map[string]string{"github": "ipetrov", "jira": "ivan.petrov@example.com"}},
		{Name: "Grace Kim", Email: "grace.kim@example.com", Manager: "Hiroshi Yamamoto", Role: "Mid-Level Engineer", TeamID: "Mobile Team",
			Identifiers: map[string]string{"github": "gkim", "jira": "grace.kim@example.com"}},

		// Data Engineering - Manager + Engineers (10 total)
		{Name: "Raj Kumar", Email: "raj.kumar@example.com", Manager: "Sarah Chen", Role: "Senior Engineer", TeamID: "Data Engineering",
			Identifiers: map[string]string{"github": "rkumar", "jira": "raj.kumar@example.com"}},
		{Name: "Chen Wang", Email: "chen.wang@example.com", Manager: "Raj Kumar", Role: "Senior Engineer", TeamID: "Data Engineering",
			Identifiers: map[string]string{"github": "cwang", "jira": "chen.wang@example.com"}},
		{Name: "Gabriela Torres", Email: "gabriela.torres@example.com", Manager: "Raj Kumar", Role: "Senior Engineer", TeamID: "Data Engineering",
			Identifiers: map[string]string{"github": "gtorres", "jira": "gabriela.torres@example.com"}},
		{Name: "Mohammed Al-Farsi", Email: "mohammed.alfarsi@example.com", Manager: "Raj Kumar", Role: "Senior Engineer", TeamID: "Data Engineering",
			Identifiers: map[string]string{"github": "malfarsi", "jira": "mohammed.alfarsi@example.com"}},
		{Name: "Anna Kowalski", Email: "anna.kowalski@example.com", Manager: "Raj Kumar", Role: "Mid-Level Engineer", TeamID: "Data Engineering",
			Identifiers: map[string]string{"github": "akowalski", "jira": "anna.kowalski@example.com"}},
		{Name: "Kwame Mensah", Email: "kwame.mensah@example.com", Manager: "Raj Kumar", Role: "Mid-Level Engineer", TeamID: "Data Engineering",
			Identifiers: map[string]string{"github": "kmensah", "jira": "kwame.mensah@example.com"}},
		{Name: "Sana Malik", Email: "sana.malik@example.com", Manager: "Raj Kumar", Role: "Mid-Level Engineer", TeamID: "Data Engineering",
			Identifiers: map[string]string{"github": "smalik", "jira": "sana.malik@example.com"}},
		{Name: "Felix Mueller", Email: "felix.mueller@example.com", Manager: "Raj Kumar", Role: "Mid-Level Engineer", TeamID: "Data Engineering",
			Identifiers: map[string]string{"github": "fmueller", "jira": "felix.mueller@example.com"}},
		{Name: "Chiara Romano", Email: "chiara.romano@example.com", Manager: "Raj Kumar", Role: "Junior Engineer", TeamID: "Data Engineering",
			Identifiers: map[string]string{"github": "cromano", "jira": "chiara.romano@example.com"}},
		{Name: "Tariq Mansour", Email: "tariq.mansour@example.com", Manager: "Raj Kumar", Role: "Junior Engineer", TeamID: "Data Engineering",
			Identifiers: map[string]string{"github": "tmansour", "jira": "tariq.mansour@example.com"}},
	}

	// Track counts by role
	roleCounts := make(map[string]int)

	// Create all engineers
	data.EngineerIDs = make([]string, len(engineers))
	for i, eng := range engineers {
		// Resolve team ID to actual team
		team, ok := data.TeamsByName[eng.TeamID]
		if !ok {
			return fmt.Errorf("team not found: %s", eng.TeamID)
		}
		actualTeamID := team.ID

		// Create engineer
		engineerID, err := data.IdentityService.CreateEngineer(eng.Name, eng.Email, eng.Manager, eng.Identifiers)
		if err != nil {
			return fmt.Errorf("failed to create engineer %s: %w", eng.Name, err)
		}
		data.EngineerIDs[i] = engineerID

		// Store engineer definition with resolved team ID
		eng.TeamID = actualTeamID
		data.EngineerMap[engineerID] = eng

		// Assign role
		roleID := data.RoleMap[eng.Role]
		_, err = data.Database.DB.Exec("UPDATE engineers SET role_id = ? WHERE id = ?", roleID, engineerID)
		if err != nil {
			return fmt.Errorf("failed to assign role to %s: %w", eng.Name, err)
		}

		// Add to team
		_, err = data.TeamStore.AddTeamMember(actualTeamID, engineerID, eng.Role)
		if err != nil {
			return fmt.Errorf("failed to add %s to team: %w", eng.Name, err)
		}

		roleCounts[eng.Role]++
	}

	fmt.Printf("\nTotal engineers created: %d\n", len(engineers))
	fmt.Println("Distribution by role:")
	for role, count := range roleCounts {
		fmt.Printf("  - %s: %d\n", role, count)
	}

	// Update team manager IDs
	sarahID := data.EngineerIDs[0] // Sarah Chen - VP Engineering
	err := data.TeamStore.UpdateTeam(data.EngineeringTeam.ID, data.EngineeringTeam.Name, nil, &sarahID)
	if err != nil {
		log.Printf("Warning: Failed to set team manager: %v", err)
	}

	return nil
}
