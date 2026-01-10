package dao

import (
	"fmt"
	"testing"
)

// Helper to create a config with N projects, M tasks, and default specs/themes/targets
func createBenchmarkConfig(numProjects, numTasks int) Config {
	config := Config{}

	// Create projects
	config.ProjectList = make([]Project, numProjects)
	for i := 0; i < numProjects; i++ {
		config.ProjectList[i] = Project{
			Name:    fmt.Sprintf("project-%d", i),
			Path:    fmt.Sprintf("/path/to/project-%d", i),
			RelPath: fmt.Sprintf("project-%d", i),
			Tags:    []string{"tag1", "tag2"},
		}
	}

	// Create tasks
	config.TaskList = make([]Task, numTasks)
	for i := 0; i < numTasks; i++ {
		config.TaskList[i] = Task{
			Name: fmt.Sprintf("task-%d", i),
			Cmd:  fmt.Sprintf("echo task %d", i),
		}
	}

	// Create specs
	config.SpecList = []Spec{
		{Name: "default", Output: "stream", Forks: 4},
		{Name: "parallel", Output: "stream", Parallel: true, Forks: 8},
	}

	// Create themes
	config.ThemeList = []Theme{
		{Name: "default"},
		{Name: "custom"},
	}

	// Create targets
	config.TargetList = []Target{
		{Name: "default", All: true},
		{Name: "frontend", Tags: []string{"frontend"}},
	}

	return config
}

// Lookup_GetProject: Find project by name (O(n) linear search)
func BenchmarkLookup_GetProject(b *testing.B) {
	sizes := []int{10, 50, 100, 500}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("projects_%d", size), func(b *testing.B) {
			config := createBenchmarkConfig(size, 10)
			// Look up a project in the middle
			targetName := fmt.Sprintf("project-%d", size/2)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = config.GetProject(targetName)
			}
		})
	}
}

// Lookup_GetTask: Find task by name (O(n) linear search)
func BenchmarkLookup_GetTask(b *testing.B) {
	sizes := []int{10, 50, 100, 500}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("tasks_%d", size), func(b *testing.B) {
			config := createBenchmarkConfig(10, size)
			// Look up a task in the middle
			targetName := fmt.Sprintf("task-%d", size/2)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = config.GetTask(targetName)
			}
		})
	}
}

// Lookup_GetSpec: Find spec by name
func BenchmarkLookup_GetSpec(b *testing.B) {
	config := createBenchmarkConfig(10, 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = config.GetSpec("default")
	}
}

// Lookup_GetTheme: Find theme by name
func BenchmarkLookup_GetTheme(b *testing.B) {
	config := createBenchmarkConfig(10, 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = config.GetTheme("default")
	}
}

// Lookup_GetTarget: Find target by name
func BenchmarkLookup_GetTarget(b *testing.B) {
	config := createBenchmarkConfig(10, 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = config.GetTarget("default")
	}
}

// Filter_ByName: Filter projects by name list
func BenchmarkFilter_ByName(b *testing.B) {
	sizes := []int{10, 50, 100}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("projects_%d", size), func(b *testing.B) {
			config := createBenchmarkConfig(size, 10)
			// Look up 5 projects
			names := []string{
				fmt.Sprintf("project-%d", size/5),
				fmt.Sprintf("project-%d", size/4),
				fmt.Sprintf("project-%d", size/3),
				fmt.Sprintf("project-%d", size/2),
				fmt.Sprintf("project-%d", size-1),
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = config.GetProjectsByName(names)
			}
		})
	}
}

// Filter_ByTags: Filter projects by tags
func BenchmarkFilter_ByTags(b *testing.B) {
	sizes := []int{10, 50, 100, 500}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("projects_%d", size), func(b *testing.B) {
			config := createBenchmarkConfig(size, 10)
			tags := []string{"tag1"}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = config.GetProjectsByTags(tags)
			}
		})
	}
}

// Filter_ByPath: Filter by path patterns (simple, *, **)
func BenchmarkFilter_ByPath(b *testing.B) {
	sizes := []int{10, 50, 100}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("projects_%d_simple", size), func(b *testing.B) {
			config := createBenchmarkConfig(size, 10)
			paths := []string{"project-1"}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = config.GetProjectsByPath(paths)
			}
		})

		b.Run(fmt.Sprintf("projects_%d_glob", size), func(b *testing.B) {
			config := createBenchmarkConfig(size, 10)
			paths := []string{"project-*"}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = config.GetProjectsByPath(paths)
			}
		})

		b.Run(fmt.Sprintf("projects_%d_doubleglob", size), func(b *testing.B) {
			config := createBenchmarkConfig(size, 10)
			paths := []string{"**/project-*"}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = config.GetProjectsByPath(paths)
			}
		})
	}
}

// Filter_Combined: FilterProjects with multiple criteria
func BenchmarkFilter_Combined(b *testing.B) {
	sizes := []int{10, 50, 100}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("projects_%d_all", size), func(b *testing.B) {
			config := createBenchmarkConfig(size, 10)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = config.FilterProjects(false, true, nil, nil, nil, "")
			}
		})

		b.Run(fmt.Sprintf("projects_%d_bytags", size), func(b *testing.B) {
			config := createBenchmarkConfig(size, 10)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = config.FilterProjects(false, false, nil, nil, []string{"tag1"}, "")
			}
		})
	}
}

// Util_ConfigLoad: Simulates config loading (ParseTask lookups)
func BenchmarkUtil_ConfigLoad(b *testing.B) {
	taskCounts := []int{10, 25, 50, 100}

	for _, numTasks := range taskCounts {
		b.Run(fmt.Sprintf("tasks_%d", numTasks), func(b *testing.B) {
			config := createBenchmarkConfig(50, numTasks)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Simulate what happens during config load:
				// Each task calls GetTheme, GetSpec, GetTarget
				for j := 0; j < numTasks; j++ {
					_, _ = config.GetTheme("default")
					_, _ = config.GetSpec("default")
					_, _ = config.GetTarget("default")
				}
			}
		})
	}
}

// Lookup_GetCommand: Find task and convert to command
func BenchmarkLookup_GetCommand(b *testing.B) {
	sizes := []int{10, 50, 100, 500}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("tasks_%d", size), func(b *testing.B) {
			config := createBenchmarkConfig(10, size)
			targetName := fmt.Sprintf("task-%d", size/2)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = config.GetCommand(targetName)
			}
		})
	}
}

// Filter_ByTagsExpr: Filter using tag expressions (&&, ||, !)
func BenchmarkFilter_ByTagsExpr(b *testing.B) {
	sizes := []int{10, 50, 100}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("projects_%d_simple", size), func(b *testing.B) {
			config := createBenchmarkConfig(size, 10)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = config.GetProjectsByTagsExpr("tag1")
			}
		})

		b.Run(fmt.Sprintf("projects_%d_and", size), func(b *testing.B) {
			config := createBenchmarkConfig(size, 10)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = config.GetProjectsByTagsExpr("tag1 && tag2")
			}
		})

		b.Run(fmt.Sprintf("projects_%d_or", size), func(b *testing.B) {
			config := createBenchmarkConfig(size, 10)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = config.GetProjectsByTagsExpr("tag1 || tag2")
			}
		})

		b.Run(fmt.Sprintf("projects_%d_complex", size), func(b *testing.B) {
			config := createBenchmarkConfig(size, 10)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = config.GetProjectsByTagsExpr("(tag1 && tag2) || !tag3")
			}
		})
	}
}

// Benchmark BuildIndices (measures index creation overhead)
// Note: Uncomment when BuildIndices optimization is added to Config
// func BenchmarkBuildIndices(b *testing.B) {
// 	sizes := []int{10, 50, 100, 500}
//
// 	for _, size := range sizes {
// 		b.Run(fmt.Sprintf("projects_%d_tasks_%d", size, size/2), func(b *testing.B) {
// 			config := createBenchmarkConfig(size, size/2)
//
// 			b.ResetTimer()
// 			for i := 0; i < b.N; i++ {
// 				config.BuildIndices()
// 			}
// 		})
// 	}
// }

// Util_GetCwdProject: Find project matching current directory
func BenchmarkUtil_GetCwdProject(b *testing.B) {
	sizes := []int{10, 50, 100, 500}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("projects_%d", size), func(b *testing.B) {
			config := createBenchmarkConfig(size, 10)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// This will search through all projects
				// In real usage, it matches against cwd
				_, _ = config.GetCwdProject()
			}
		})
	}
}

// Filter_Intersect: Intersection of project lists
func BenchmarkFilter_Intersect(b *testing.B) {
	sizes := []int{10, 50, 100}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("projects_%d", size), func(b *testing.B) {
			config := createBenchmarkConfig(size, 10)
			// Create two overlapping project lists
			list1 := config.ProjectList[:size/2]
			list2 := config.ProjectList[size/4:]

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = config.GetIntersectProjects(list1, list2)
			}
		})
	}
}
